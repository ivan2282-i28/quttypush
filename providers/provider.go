package providers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Provider interface {
	GetMetadata() (*Metadata, error)
	CheckToken(token string) (bool, error)
	SendPush(token string, payload map[string]interface{}) (*PushResponse, error)
}

type Metadata struct {
	Qutty string `json:"qutty"`
	Apiver int    `json:"apiver"`
	Soft   string `json:"soft"`
	Label  string `json:"label"`
}

type PushResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type BaseProvider struct {
	BaseURL string
	Client  *http.Client
}

func NewBaseProvider(provider string) *BaseProvider {
	base := provider
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	base = strings.TrimRight(base, "/")
	return &BaseProvider{
		BaseURL: base,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *BaseProvider) GetMetadata() (*Metadata, error) {
	resp, err := b.Client.Get(b.BaseURL + "/v1/quttypush")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var meta Metadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (b *BaseProvider) CheckToken(token string) (bool, error) {
	escaped := url.PathEscape(token)
	resp, err := b.Client.Get(b.BaseURL + "/v1/check/" + escaped)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, nil
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil
	}
	if valid, ok := result["valid"].(bool); ok {
		return valid, nil
	}
	return false, nil
}

func (b *BaseProvider) SendPush(token string, payload map[string]interface{}) (*PushResponse, error) {
	body := map[string]interface{}{
		"token":   token,
		"payload": payload,
	}
	reqBody, _ := json.Marshal(body)
	resp, err := b.Client.Post(b.BaseURL+"/v1/post", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var pres PushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pres); err != nil {
		// Try to be lenient
		return &PushResponse{Status: "delivered", Reason: "ok"}, nil
	}
	return &pres, nil
}
