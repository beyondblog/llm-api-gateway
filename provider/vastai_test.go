package provider

import (
	"os"
	"testing"
)

func TestVastAIProvider_GetEndpoints(t *testing.T) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	provider := NewVastAIProvider(apiKey, "TheBloke/Mixtral-8x7B-Instruct-v0.1-GPTQ", "main", "")
	endpoints, err := provider.GetEndpoints()
	if err != nil {
		t.Error(err)
	}
	t.Logf("endpoints: %v", endpoints)
}

func TestVastAI_ExecuteCommand(t *testing.T) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	provider := NewVastAIProvider(apiKey, "TheBloke/Mixtral-8x7B-Instruct-v0.1-GPTQ", "main", "")
	instanceID := 7860040
	r, err := provider.executeCommand(instanceID, "ls -llah /app")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("response: %v", r)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("response: %v", r)
}

func TestVastAIProvider_AutoScaling(t *testing.T) {
	apiKey := os.Getenv("VASTAI_API_KEY")
	provider := NewVastAIProvider(apiKey, "TheBloke/Mixtral-8x7B-Instruct-v0.1-GPTQ", "main", "")
	_ = provider.createInstance()
}
