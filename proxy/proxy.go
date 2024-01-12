package proxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/beyondblog/llm-api-gateway/provider"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
)

// Backend holds the data about a server
type Backend struct {
	URL            *url.URL
	Alive          bool
	mux            sync.RWMutex
	ReverseProxy   *httputil.ReverseProxy
	HealthCheckURL string
}

// SetAlive for this backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// IsAlive returns true when backend is alive
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

// ServerPool holds information about reachable backends
type ServerPool struct {
	backends []*Backend
	current  uint64
	close    bool
}

func (s *ServerPool) Destroy() {
	s.close = true
}

func (s *ServerPool) IsClose() bool {
	return s.close
}

// AddBackend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

// NextIndex atomically increase the counter and return an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// MarkBackendStatus changes a status of a backend
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// GetNextPeer returns next active peer to take a connection
func (s *ServerPool) GetNextPeer() *Backend {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends)     // take an index by modding
		if s.backends[idx].IsAlive() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

// HealthCheck pings the backends and update the status
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := httpHealthCheck(fmt.Sprintf("%s%s", b.URL.String(), b.HealthCheckURL))
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

// GetAttemptsFromContext returns the attempts for request
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

// GetRetryFromContext returns the retries for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

// isAlive checks whether a backend is Alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return true
}

func httpHealthCheck(url string) bool {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	if response != nil {
		defer func() {
			_ = response.Body.Close()
		}()
	}
	return true
}

// healthCheck runs a routine for check status of the backends every 3 mins
func healthCheck(serverPool *ServerPool) {
	serverPool.HealthCheck()
	t := time.NewTicker(time.Minute * 3)
	for {
		select {
		case <-t.C:
			if serverPool.IsClose() {
				return
			}
			log.Println("Starting health check...")
			serverPool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

type Server struct {
	BackendList []string
	serverPool  *ServerPool
	llmProvider provider.LLMProvider
}

func NewProxyServer(llmProvider provider.LLMProvider) *Server {
	server := new(Server)
	server.llmProvider = llmProvider
	return server
}

func (s *Server) ReloadBackend() {
	if s.serverPool != nil {
		s.serverPool.Destroy()
	}
	s.serverPool = new(ServerPool)
	serverEndpoint, err := s.llmProvider.GetEndpoints()
	if err != nil {
		log.Fatal(err)
	}

	if len(serverEndpoint) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	log.Printf("Loading endpoints size: %d\n", len(serverEndpoint))
	// parse servers
	for _, endpoint := range serverEndpoint {
		//goland:noinspection HttpUrlsUsage
		serverUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", endpoint.Host, endpoint.Port))
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {

			if errors.Is(e, context.Canceled) {
				writer.WriteHeader(http.StatusBadGateway)
				return
			}

			log.Printf("[%s] %s %s\n", serverUrl.Host, request.URL.RequestURI(), e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this backend as down
			s.serverPool.MarkBackendStatus(serverUrl, false)

			// if the same request routing for few attempts with different backends, increase the count
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			s.lb(writer, request.WithContext(ctx))
		}

		s.serverPool.AddBackend(&Backend{
			URL:            serverUrl,
			Alive:          true,
			ReverseProxy:   proxy,
			HealthCheckURL: "/v1/models",
		})
		log.Printf("host %s found\n", serverUrl)
	}

	// start health checking
	go healthCheck(s.serverPool)
}

func (s *Server) Run(port int) {
	// load backends
	s.ReloadBackend()
	// create http server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(s.lb),
	}

	go s.SyncBackend()
	log.Printf("Load Balancer started at :%d\n", port)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	// Waiting for SIGINT (kill -2)
	<-stop
	log.Printf("Stop server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// handle err
	}

}

// lb load balances the incoming request
func (s *Server) lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := s.serverPool.GetNextPeer()
	if peer != nil {
		startTime := time.Now()
		peer.ReverseProxy.ServeHTTP(w, r)
		since := time.Since(startTime)

		s.logRequest(r, since, peer.URL.Host)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (s *Server) logRequest(req *http.Request, elapsedTime time.Duration, backend string) {
	clientIP := req.RemoteAddr
	elapsedTimeFormatted := fmt.Sprintf("%.3f", elapsedTime.Seconds())
	log.Printf("%s - [%s] \"%s %s\"  %s %s\n", clientIP, req.Method, req.URL.RequestURI(), req.Proto,
		backend, elapsedTimeFormatted)
}

// SyncBackend Scheduled synchronization of backend
func (s *Server) SyncBackend() {
	t := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-t.C:
			s.ReloadBackend()
		}
	}
}
