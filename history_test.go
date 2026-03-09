package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNeedName_AllValid(t *testing.T) {
	expected := []string{"Hunger", "Comfort", "Bladder", "Energy", "Fun", "Social", "Hygiene", "Environment"}
	for i, want := range expected {
		got := NeedName(Need(i))
		if got != want {
			t.Errorf("NeedName(%d) = %q, want %q", i, got, want)
		}
	}
}

func TestNeedName_OutOfBounds(t *testing.T) {
	for _, n := range []Need{-1, NeedCount, 100} {
		got := NeedName(n)
		if got != "Unknown" {
			t.Errorf("NeedName(%d) = %q, want %q", n, got, "Unknown")
		}
	}
}

func TestAppendEntry_Empty(t *testing.T) {
	h := History{}
	e := Entry{Timestamp: time.Now(), Values: [NeedCount]int{5, 5, 5, 5, 5, 5, 5, 5}}

	result := AppendEntry(h, e)

	if len(result.Entries) != 1 {
		t.Fatalf("len = %d, want 1", len(result.Entries))
	}
	if result.Entries[0].Values != e.Values {
		t.Errorf("values mismatch")
	}
}

func TestAppendEntry_Preserves(t *testing.T) {
	e1 := Entry{Timestamp: time.Now(), Values: [NeedCount]int{1, 2, 3, 4, 5, 6, 7, 8}}
	h := History{Entries: []Entry{e1}}
	e2 := Entry{Timestamp: time.Now(), Values: [NeedCount]int{8, 7, 6, 5, 4, 3, 2, 1}}

	result := AppendEntry(h, e2)

	if len(result.Entries) != 2 {
		t.Fatalf("len = %d, want 2", len(result.Entries))
	}
	if result.Entries[0].Values != e1.Values {
		t.Error("first entry changed")
	}
	if result.Entries[1].Values != e2.Values {
		t.Error("second entry mismatch")
	}
	// Original unchanged
	if len(h.Entries) != 1 {
		t.Error("original history mutated")
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	h := History{Entries: []Entry{
		{Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), Values: [NeedCount]int{1, 2, 3, 4, 5, 6, 7, 8}},
		{Timestamp: time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC), Values: [NeedCount]int{8, 7, 6, 5, 4, 3, 2, 1}},
	}}

	if err := Save(path, h); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Entries) != len(h.Entries) {
		t.Fatalf("len = %d, want %d", len(loaded.Entries), len(h.Entries))
	}
	for i := range h.Entries {
		if loaded.Entries[i].Values != h.Entries[i].Values {
			t.Errorf("entry %d values mismatch", i)
		}
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	h, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected empty history, got %d entries", len(h.Entries))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.json")

	err := Save(path, History{})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}
}
