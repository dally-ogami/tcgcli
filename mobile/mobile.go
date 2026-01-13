package tcgmobile

import (
	"encoding/json"
	"errors"
	"fmt"

	"tcgcli/tcg"
)

type Manager struct {
	manager *tcg.DeckManager
	deck    *tcg.Deck
}

type DeckStatus struct {
	DeckName       string             `json:"deck_name"`
	DeckPath       string             `json:"deck_path"`
	CardsSource    tcg.CardsSource    `json:"cards_source"`
	LoadStatus     tcg.DeckLoadStatus `json:"load_status"`
	CardsLoadError string             `json:"cards_load_error,omitempty"`
}

func NewManager(decksDir string) (*Manager, error) {
	manager, err := tcg.NewDeckManager(decksDir)
	if err != nil {
		return nil, err
	}
	return &Manager{manager: manager}, nil
}

func (m *Manager) ListDecksJSON() (string, error) {
	decks, err := m.manager.ListExistingDecks()
	if err != nil {
		return "", err
	}
	return toJSON(decks)
}

func (m *Manager) CreateDeck(name string) error {
	deck, err := m.manager.CreateDeck(name)
	if err != nil {
		return err
	}
	m.deck = deck
	return nil
}

func (m *Manager) LoadDeck(name string) error {
	deck, err := m.manager.LoadDeck(name)
	if err != nil {
		return err
	}
	m.deck = deck
	return nil
}

func (m *Manager) DeckStatusJSON() (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	status := DeckStatus{
		DeckName:    deck.Name,
		DeckPath:    deck.FilePath,
		CardsSource: deck.CardsSource,
		LoadStatus:  deck.LoadStatus,
	}
	if deck.CardsLoadError != nil {
		status.CardsLoadError = deck.CardsLoadError.Error()
	}
	return toJSON(status)
}

func (m *Manager) AvailableCardsJSON() (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	return toJSON(deck.ListAvailableCards())
}

func (m *Manager) SearchCardsJSON(term string) (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	return toJSON(deck.SearchCards(term))
}

func (m *Manager) DeckCardsJSON() (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	return toJSON(deck.Cards)
}

func (m *Manager) AddCardByIDJSON(cardID string) (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	result, err := deck.AddCardByID(cardID)
	if err != nil {
		return "", err
	}
	return toJSON(result)
}

func (m *Manager) RemoveCardJSON(index int) (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	entry, err := deck.RemoveCard(index)
	if err != nil {
		return "", err
	}
	return toJSON(entry)
}

func (m *Manager) RecordBattle(result, opponent string) error {
	deck, err := m.currentDeck()
	if err != nil {
		return err
	}
	return deck.RecordBattle(result, opponent, tcg.Now())
}

func (m *Manager) StatsJSON() (string, error) {
	deck, err := m.currentDeck()
	if err != nil {
		return "", err
	}
	return toJSON(deck.Stats())
}

func (m *Manager) SaveDeck() error {
	deck, err := m.currentDeck()
	if err != nil {
		return err
	}
	return deck.Save()
}

func (m *Manager) currentDeck() (*tcg.Deck, error) {
	if m.deck == nil {
		return nil, errors.New("no deck loaded")
	}
	return m.deck, nil
}

func toJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode JSON: %w", err)
	}
	return string(payload), nil
}
