package main

import (
	"flag"
	"github.com/beyondblog/llm-api-gateway/provider"
	"github.com/beyondblog/llm-api-gateway/proxy"
)

var (
	vastAIAPIKey = flag.String("vastai_api_key", "", "vast.ai api key")
)

func main() {
	flag.Parse()
	vastAIProvider := provider.NewVastAIProvider(*vastAIAPIKey)
	proxyServer := proxy.NewProxyServer(vastAIProvider)
	proxyServer.Run(8888)
}
