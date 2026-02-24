package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestManifestParsing(t *testing.T) {
	tmpDir := t.TempDir()
	m := Manifest{
		Name:       "test-plugin",
		Type:       "provider",
		Executable: "test-binary",
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(tmpDir, "plugin.yaml")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Read it back
	readData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Manifest
	if err := yaml.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Name != "test-plugin" {
		t.Errorf("name = %q, want %q", loaded.Name, "test-plugin")
	}
	if loaded.Type != "provider" {
		t.Errorf("type = %q, want %q", loaded.Type, "provider")
	}
}

func TestJSONRPCRequestMarshal(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "test",
		Params:  map[string]string{"key": "value"},
	}

	if req.Method != "test" {
		t.Errorf("method = %q", req.Method)
	}
	if req.ID != 1 {
		t.Errorf("id = %d", req.ID)
	}
}

func TestRPCError(t *testing.T) {
	err := &RPCError{Code: -32601, Message: "method not found"}
	if err.Error() != "RPC error -32601: method not found" {
		t.Errorf("error = %q", err.Error())
	}
}
