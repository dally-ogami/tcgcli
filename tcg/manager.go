package tcg

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DeckManager struct {
	DecksDir string
}

func NewDeckManager(decksDir string) (*DeckManager, error) {
	if err := os.MkdirAll(decksDir, 0o755); err != nil {
		return nil, err
	}
	return &DeckManager{DecksDir: decksDir}, nil
}

func (m *DeckManager) ListExistingDecks() ([]string, error) {
	entries, err := os.ReadDir(m.DecksDir)
	if err != nil {
		return nil, err
	}

	var decks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			decks = append(decks, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(decks)
	return decks, nil
}

func (m *DeckManager) CreateDeck(name string) (*Deck, error) {
	deckFile := filepath.Join(m.DecksDir, name+".json")
	if _, err := os.Stat(deckFile); err == nil {
		return nil, os.ErrExist
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return NewDeck(name, deckFile)
}

func (m *DeckManager) LoadDeck(name string) (*Deck, error) {
	deckFile := filepath.Join(m.DecksDir, name+".json")
	return NewDeck(name, deckFile)
}
