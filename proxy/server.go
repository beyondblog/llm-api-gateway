package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/beyondblog/llm-api-gateway/provider"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type fakeCloseReadCloser struct {
	io.ReadCloser
}

func (w *fakeCloseReadCloser) Close() error {
	return nil
}

func (w *fakeCloseReadCloser) RealClose() error {
	if w.ReadCloser == nil {
		return nil
	}
	return w.ReadCloser.Close()
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

	serverPool := new(ServerPool)
	serverEndpoint, err := s.llmProvider.GetEndpoints()
	if err != nil {
		log.Printf("GetEndpoints err: %v\n", err)
		return
	}

	if len(serverEndpoint) == 0 {
		log.Printf("Please provide one or more backends to load balance")
		return
	}

	log.Printf("Loading endpoints size: %d\n", len(serverEndpoint))
	for _, endpoint := range serverEndpoint {
		//goland:noinspection HttpUrlsUsage
		serverUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", endpoint.Host, endpoint.Port))
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {

			statusCode := http.StatusInternalServerError
			if e, ok := err.(net.Error); ok {
				if e.Timeout() {
					statusCode = http.StatusGatewayTimeout
				} else {
					statusCode = http.StatusBadGateway
				}
			} else if err == io.EOF {
				statusCode = http.StatusBadGateway
			} else if errors.Is(err, context.Canceled) {
				statusCode = 499
				writer.WriteHeader(statusCode)
				return
			}

			log.Printf("[%s] %s %s\n", serverUrl.Host, request.URL.RequestURI(), err.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)

					if request.GetBody != nil {
						b, _ := request.GetBody()
						request.Body = b
					}
					proxy.ServeHTTP(writer, request.WithContext(ctx))
					if request.Body != nil {
						_, ok := request.Body.(*fakeCloseReadCloser)
						if ok {
							_ = request.Body.(*fakeCloseReadCloser).RealClose()
						}
					}
				}
				return
			}

			// after 3 retries, mark this backend as down
			serverPool.MarkBackendStatus(serverUrl, false)
			log.Printf("%s [%s]\n", serverUrl.String(), "down")

			// if the same request routing for few attempts with different backends, increase the count
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			r := request.WithContext(ctx)
			if r.GetBody != nil {
				b, _ := r.GetBody()
				r.Body = b
			}
			s.lb(writer, r)
			if r.Body != nil {
				_, ok := r.Body.(*fakeCloseReadCloser)
				if ok {
					_ = r.Body.(*fakeCloseReadCloser).RealClose()
				}
			}
		}

		serverPool.AddBackend(&Backend{
			URL:            serverUrl,
			Alive:          true,
			ReverseProxy:   proxy,
			HealthCheckURL: "/v1/internal/model/info",
		})
		log.Printf("host %s found\n", serverUrl)
	}

	if s.serverPool != nil {
		s.serverPool.Destroy()
	}

	s.serverPool = serverPool
	s.serverPool.HealthCheck()
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
	log.Printf("Model %s", s.llmProvider.GetModel())
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

	if peer == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	startTime := time.Now()

	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = &fakeCloseReadCloser{io.NopCloser(bytes.NewBuffer(bodyBytes))}
		r.GetBody = func() (io.ReadCloser, error) {
			body := io.NopCloser(bytes.NewBuffer(bodyBytes))
			return body, nil
		}

		defer func() {
			if r.Body != nil {
				return
			}
			_, ok := r.Body.(*fakeCloseReadCloser)
			if ok {
				_ = r.Body.(*fakeCloseReadCloser).RealClose()
			}
		}()

	}

	peer.ReverseProxy.ServeHTTP(w, r)
	since := time.Since(startTime)

	s.logRequest(r, since, peer.URL.Host)
	return
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
