package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	. "q/types"
)

// PluginProvider implements the Provider interface by delegating to a plugin process.
type PluginProvider struct {
	process *Process
}

// NewPluginProvider starts the plugin process and returns a PluginProvider.
func NewPluginProvider(executablePath string) (*PluginProvider, error) {
	proc, err := StartProcess(executablePath)
	if err != nil {
		return nil, err
	}
	return &PluginProvider{process: proc}, nil
}

// SetHeaders calls the plugin's set_headers method and applies returned headers.
func (p *PluginProvider) SetHeaders(req *http.Request, auth, orgID string) {
	params := map[string]string{"auth": auth, "org_id": orgID}
	result, err := p.process.Call("set_headers", params)
	if err != nil {
		return
	}
	var headers map[string]string
	if err := json.Unmarshal(result, &headers); err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// BuildRequestBody calls the plugin's build_request_body method.
func (p *PluginProvider) BuildRequestBody(payload Payload) (interface{}, error) {
	result, err := p.process.Call("build_request_body", payload)
	if err != nil {
		return nil, fmt.Errorf("plugin build_request_body failed: %w", err)
	}
	var body interface{}
	if err := json.Unmarshal(result, &body); err != nil {
		return nil, err
	}
	return body, nil
}

// ParseStreamLine calls the plugin's parse_stream_line method.
func (p *PluginProvider) ParseStreamLine(line, eventType string) (string, bool, bool) {
	params := map[string]string{"line": line, "event_type": eventType}
	result, err := p.process.Call("parse_stream_line", params)
	if err != nil {
		return "", false, true
	}
	var resp struct {
		Content    string `json:"content"`
		IsDone     bool   `json:"is_done"`
		ShouldSkip bool   `json:"should_skip"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", false, true
	}
	return resp.Content, resp.IsDone, resp.ShouldSkip
}

// Stop terminates the plugin process.
func (p *PluginProvider) Stop() {
	if p.process != nil {
		p.process.Stop()
	}
}
