package theme

import "testing"

func TestLoadTheme(t *testing.T) {
	for _, name := range Names() {
		if !LoadTheme(name) {
			t.Errorf("LoadTheme(%q) returned false", name)
		}
		if Current.Name != name {
			t.Errorf("Current.Name = %q, want %q", Current.Name, name)
		}
	}
}

func TestLoadThemeInvalid(t *testing.T) {
	if LoadTheme("nonexistent") {
		t.Error("expected false for nonexistent theme")
	}
}

func TestNamesCount(t *testing.T) {
	names := Names()
	if len(names) != 6 {
		t.Errorf("expected 6 themes, got %d", len(names))
	}
}
