package history

import (
	"path/filepath"
	. "q/types"
	"testing"
	"time"
)

func testStore(t *testing.T) *HistoryStore {
	t.Helper()
	return NewStoreWithPath(filepath.Join(t.TempDir(), "history.jsonl"))
}

func TestSaveAndList(t *testing.T) {
	s := testStore(t)

	for i := 0; i < 5; i++ {
		err := s.Save(Conversation{
			ID:        string(rune('a' + i)),
			Timestamp: time.Now(),
			Model:     "test",
			Messages:  []Message{{Role: "user", Content: "hello"}},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	convs, err := s.List(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(convs) != 3 {
		t.Fatalf("expected 3 conversations, got %d", len(convs))
	}
	// Most recent first
	if convs[0].ID != "e" {
		t.Errorf("expected most recent first, got ID=%q", convs[0].ID)
	}
}

func TestSearch(t *testing.T) {
	s := testStore(t)

	s.Save(Conversation{
		ID: "1", Timestamp: time.Now(), Model: "test",
		Messages: []Message{{Role: "user", Content: "list files in directory"}},
	})
	s.Save(Conversation{
		ID: "2", Timestamp: time.Now(), Model: "test",
		Messages: []Message{{Role: "user", Content: "what is docker"}},
	})

	results, err := s.Search("files")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Errorf("expected ID=1, got %q", results[0].ID)
	}
}

func TestShow(t *testing.T) {
	s := testStore(t)

	s.Save(Conversation{ID: "abc", Timestamp: time.Now(), Model: "test",
		Messages: []Message{{Role: "user", Content: "hi"}}})

	conv, err := s.Show("abc")
	if err != nil {
		t.Fatal(err)
	}
	if conv.ID != "abc" {
		t.Errorf("ID = %q", conv.ID)
	}

	_, err = s.Show("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

func TestClear(t *testing.T) {
	s := testStore(t)
	s.Save(Conversation{ID: "1", Timestamp: time.Now(), Model: "test",
		Messages: []Message{{Role: "user", Content: "hi"}}})

	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}

	convs, _ := s.List(10)
	if len(convs) != 0 {
		t.Errorf("expected 0 after clear, got %d", len(convs))
	}
}

func TestPrune(t *testing.T) {
	s := testStore(t)

	s.Save(Conversation{ID: "old", Timestamp: time.Now().AddDate(0, 0, -60), Model: "test",
		Messages: []Message{{Role: "user", Content: "old"}}})
	s.Save(Conversation{ID: "new", Timestamp: time.Now(), Model: "test",
		Messages: []Message{{Role: "user", Content: "new"}}})

	if err := s.Prune(30); err != nil {
		t.Fatal(err)
	}

	convs, _ := s.List(10)
	if len(convs) != 1 || convs[0].ID != "new" {
		t.Errorf("expected only 'new' conversation, got %d", len(convs))
	}
}

func TestListEmpty(t *testing.T) {
	s := testStore(t)
	convs, err := s.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(convs) != 0 {
		t.Errorf("expected 0, got %d", len(convs))
	}
}
