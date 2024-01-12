package provider

import (
	"os"
	"testing"
)

func TestVastAIProvider_GetEndpoints(t *testing.T) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	provider := NewVastAIProvider(apiKey)
	endpoints, err := provider.GetEndpoints()
	if err != nil {
		t.Error(err)
	}
	t.Logf("endpoints: %v", endpoints)
}
