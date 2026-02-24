package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const pluginDir = ".shelly-ai/plugins"
const rpcTimeout = 5 * time.Second

// Manifest describes a plugin's metadata.
type Manifest struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"` // "provider"
	Executable string `yaml:"executable"`
}

// JSONRPCRequest is a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Process manages a running plugin process.
type Process struct {
	cmd    *exec.Cmd
	stdin  *json.Encoder
	stdout *bufio.Scanner
	mu     sync.Mutex
	nextID int
}

// StartProcess launches a plugin executable and returns a Process handle.
func StartProcess(executablePath string) (*Process, error) {
	cmd := exec.Command(executablePath) //nolint:gosec
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	return &Process{
		cmd:    cmd,
		stdin:  json.NewEncoder(stdinPipe),
		stdout: bufio.NewScanner(stdoutPipe),
		nextID: 1,
	}, nil
}

// Call sends a JSON-RPC request and waits for a response.
func (p *Process) Call(method string, params interface{}) (json.RawMessage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.nextID
	p.nextID++

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := p.stdin.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}

	// Read response with timeout
	done := make(chan struct{})
	var resp JSONRPCResponse
	var scanErr error

	go func() {
		if p.stdout.Scan() {
			scanErr = json.Unmarshal(p.stdout.Bytes(), &resp)
		} else {
			scanErr = fmt.Errorf("plugin process closed stdout")
		}
		close(done)
	}()

	select {
	case <-done:
		if scanErr != nil {
			return nil, scanErr
		}
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	case <-time.After(rpcTimeout):
		return nil, fmt.Errorf("plugin RPC call timed out after %v", rpcTimeout)
	}
}

// Stop terminates the plugin process.
func (p *Process) Stop() {
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
	}
}

// Discover finds all plugins in the plugin directory.
func Discover() ([]Manifest, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(homeDir, pluginDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var manifests []Manifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(dir, entry.Name(), "plugin.yaml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var m Manifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			continue
		}
		// Make executable path absolute
		if !filepath.IsAbs(m.Executable) {
			m.Executable = filepath.Join(dir, entry.Name(), m.Executable)
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}
