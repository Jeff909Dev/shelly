package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "q/types"
)

const historyFile = ".shelly-ai/history.jsonl"

// Conversation represents a single query/response interaction.
type Conversation struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
}

// HistoryStore manages conversation persistence.
type HistoryStore struct {
	filePath string
}

// NewStore creates a HistoryStore with the default file path.
func NewStore() (*HistoryStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fp := filepath.Join(homeDir, historyFile)
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &HistoryStore{filePath: fp}, nil
}

// NewStoreWithPath creates a HistoryStore with a custom file path (for testing).
func NewStoreWithPath(path string) *HistoryStore {
	return &HistoryStore{filePath: path}
}

// Save appends a conversation to the history file.
func (s *HistoryStore) Save(conv Conversation) error {
	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(conv)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// List returns the most recent n conversations.
func (s *HistoryStore) List(n int) ([]Conversation, error) {
	all, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	if n > len(all) {
		n = len(all)
	}
	// Return most recent first
	result := make([]Conversation, n)
	for i := 0; i < n; i++ {
		result[i] = all[len(all)-1-i]
	}
	return result, nil
}

// Search finds conversations containing the query string in any message.
func (s *HistoryStore) Search(query string) ([]Conversation, error) {
	all, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(query)
	var results []Conversation
	for _, conv := range all {
		for _, msg := range conv.Messages {
			if strings.Contains(strings.ToLower(msg.Content), query) {
				results = append(results, conv)
				break
			}
		}
	}
	return results, nil
}

// Show returns a specific conversation by ID.
func (s *HistoryStore) Show(id string) (*Conversation, error) {
	all, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	for _, conv := range all {
		if conv.ID == id {
			return &conv, nil
		}
	}
	return nil, fmt.Errorf("conversation %q not found", id)
}

// Clear removes all history.
func (s *HistoryStore) Clear() error {
	return os.Remove(s.filePath)
}

// Prune removes conversations older than maxDays.
func (s *HistoryStore) Prune(maxDays int) error {
	all, err := s.loadAll()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cutoff := time.Now().AddDate(0, 0, -maxDays)
	var kept []Conversation
	for _, conv := range all {
		if conv.Timestamp.After(cutoff) {
			kept = append(kept, conv)
		}
	}
	return s.writeAll(kept)
}

func (s *HistoryStore) loadAll() ([]Conversation, error) {
	f, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var convs []Conversation
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var conv Conversation
		if err := json.Unmarshal([]byte(line), &conv); err != nil {
			continue // skip corrupted lines
		}
		convs = append(convs, conv)
	}
	return convs, scanner.Err()
}

func (s *HistoryStore) writeAll(convs []Conversation) error {
	f, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, conv := range convs {
		data, err := json.Marshal(conv)
		if err != nil {
			continue
		}
		f.Write(append(data, '\n'))
	}
	return nil
}
