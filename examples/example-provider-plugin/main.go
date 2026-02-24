// Example provider plugin for Shelly AI.
// This plugin echoes back requests for demonstration purposes.
// Build: go build -o example-provider-plugin .
// Install: cp example-provider-plugin ~/.shelly-ai/plugins/example-echo/
// Also copy plugin.yaml to the same directory.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		var result interface{}

		switch req.Method {
		case "set_headers":
			// Return headers to set on the HTTP request
			result = map[string]string{
				"Authorization": "Bearer example-key",
			}

		case "build_request_body":
			// Pass through the payload as-is
			var payload interface{}
			json.Unmarshal(req.Params, &payload)
			result = payload

		case "parse_stream_line":
			var params struct {
				Line      string `json:"line"`
				EventType string `json:"event_type"`
			}
			json.Unmarshal(req.Params, &params)

			if params.Line == "data: [DONE]" {
				result = map[string]interface{}{
					"content":     "",
					"is_done":     true,
					"should_skip": false,
				}
			} else {
				result = map[string]interface{}{
					"content":     fmt.Sprintf("[echo] %s", params.Line),
					"is_done":     false,
					"should_skip": false,
				}
			}

		default:
			encoder.Encode(response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   map[string]interface{}{"code": -32601, "message": "method not found"},
			})
			continue
		}

		encoder.Encode(response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		})
	}
}
