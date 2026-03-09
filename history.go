package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Need int

const (
	Hunger Need = iota
	Comfort
	Bladder
	Energy
	Fun
	Social
	Hygiene
	Environment
	NeedCount
)

var needNames = [NeedCount]string{
	"Hunger",
	"Comfort",
	"Bladder",
	"Energy",
	"Fun",
	"Social",
	"Hygiene",
	"Environment",
}

func NeedName(n Need) string {
	if n < 0 || n >= NeedCount {
		return "Unknown"
	}
	return needNames[n]
}

func NeedNames() [NeedCount]string {
	return needNames
}

type Entry struct {
	Timestamp time.Time      `json:"timestamp"`
	Values    [NeedCount]int `json:"values"`
}

type History struct {
	Entries []Entry `json:"entries"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sbars.json"), nil
}

func Load(path string) (History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return History{}, nil
		}
		return History{}, err
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return History{}, err
	}
	return h, nil
}

func Save(path string, h History) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func AppendEntry(h History, e Entry) History {
	newEntries := make([]Entry, len(h.Entries), len(h.Entries)+1)
	copy(newEntries, h.Entries)
	newEntries = append(newEntries, e)
	return History{Entries: newEntries}
}
