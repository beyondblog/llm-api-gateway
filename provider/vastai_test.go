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

func TestVastAI_ExecuteCommand(t *testing.T) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	provider := NewVastAIProvider(apiKey)
	instanceID := 7860040
	r, err := provider.executeCommand(instanceID, "ls -llah /app")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("response: %v", r)
}
