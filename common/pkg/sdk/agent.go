package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iat/common/protocol"
	"net/http"
	"time"
)

type AgentConfig struct {
	ID           uint                  `json:"id"`
	Name         string                `json:"name"`
	RegistryURL  string                `json:"registryUrl"`
	Capabilities []protocol.Capability `json:"capabilities"`
	Endpoint     string                `json:"endpoint"`
}

type AgentSDK struct {
	config AgentConfig
	client *http.Client
}

func NewAgentSDK(config AgentConfig) *AgentSDK {
	return &AgentSDK{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (sdk *AgentSDK) Register() error {
	reqBody := map[string]any{
		"agentId":      sdk.config.ID,
		"capabilities": sdk.config.Capabilities,
		"endpoint":     sdk.config.Endpoint,
	}
	data, _ := json.Marshal(reqBody)
	resp, err := sdk.client.Post(sdk.config.RegistryURL+"/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}
	return nil
}

func (sdk *AgentSDK) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sdk.sendHeartbeat()
			}
		}
	}()
}

func (sdk *AgentSDK) sendHeartbeat() {
	url := fmt.Sprintf("%s/heartbeat/%d", sdk.config.RegistryURL, sdk.config.ID)
	resp, err := sdk.client.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("Heartbeat failed: %v\n", err)
		return
	}
	resp.Body.Close()
}
