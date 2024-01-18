package main

import (
	"flag"
	"github.com/beyondblog/llm-api-gateway/provider"
	"github.com/beyondblog/llm-api-gateway/proxy"
)

var (
	vastAIAPIKey = flag.String("vastai_api_key", "", "vast.ai api key")
	port         = flag.Int("port", 8080, "Port to serve")
	model        = flag.String("model", "gpt2", "model name")
	branch       = flag.String("branch", "main", "branch name")
)

func main() {
	flag.Parse()
	vastAIProvider := provider.NewVastAIProvider(*vastAIAPIKey, *model, *branch)
	proxyServer := proxy.NewProxyServer(vastAIProvider)
	proxyServer.Run(*port)
}
