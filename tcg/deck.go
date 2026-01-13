package tcg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Deck struct {
	Name           string
	FilePath       string
	Cards          []CardEntry
	BattleHistory  []BattleRecord
	ValidCards     []Card
	CardsSource    CardsSource
	CardsLoadError error
	LoadStatus     DeckLoadStatus
}

func NewDeck(name, filePath string) (*Deck, error) {
	deck := &Deck{
		Name:     name,
		FilePath: filePath,
	}

	cards, source, warn, err := LoadValidCards()
	if err != nil {
		return nil, err
	}
	deck.ValidCards = cards
	deck.CardsSource = source
	deck.CardsLoadError = warn

	status, err := deck.loadDeckFile()
	if err != nil {
		return nil, err
	}
	deck.LoadStatus = status

	return deck, nil
}

func (d *Deck) loadDeckFile() (DeckLoadStatus, error) {
	if _, err := os.Stat(d.FilePath); errors.Is(err, os.ErrNotExist) {
		d.Cards = []CardEntry{}
		d.BattleHistory = []BattleRecord{}
		return DeckLoadNew, nil
	} else if err != nil {
		return DeckLoadReset, err
	}

	file, err := os.Open(d.FilePath)
	if err != nil {
		return DeckLoadReset, err
	}
	defer file.Close()

	var data deckFileData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		d.Cards = []CardEntry{}
		d.BattleHistory = []BattleRecord{}
		return DeckLoadReset, nil
	}

	d.Cards = data.Cards
	d.BattleHistory = data.BattleHistory
	return DeckLoadLoaded, nil
}

func (d *Deck) Save() error {
	data := deckFileData{
		Cards:         d.Cards,
		BattleHistory: d.BattleHistory,
	}

	if err := os.MkdirAll(filepath.Dir(d.FilePath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(d.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(&data)
}

func (d *Deck) ListAvailableCards() []Card {
	return append([]Card(nil), d.ValidCards...)
}

func (d *Deck) SearchCards(term string) []Card {
	normalized := strings.ToLower(strings.TrimSpace(term))
	if normalized == "" {
		return nil
	}

	var matches []Card
	for _, card := range d.ValidCards {
		name := strings.ToLower(card.Name)
		set := strings.ToLower(card.Set)
		if strings.Contains(name, normalized) || strings.Contains(set, normalized) {
			matches = append(matches, card)
		}
	}
	return matches
}

func (d *Deck) FindCardByID(cardID string) (Card, bool) {
	needle := strings.ToLower(strings.TrimSpace(cardID))
	if needle == "" {
		return Card{}, false
	}
	for _, card := range d.ValidCards {
		if strings.EqualFold(card.ID, needle) {
			return card, true
		}
	}
	return Card{}, false
}

func (d *Deck) AddCardByID(cardID string) (AddCardResult, error) {
	card, ok := d.FindCardByID(cardID)
	if !ok {
		return AddCardResult{}, fmt.Errorf("card ID %q not found", cardID)
	}

	cardName := strings.TrimSpace(card.Name)
	cardSet := strings.TrimSpace(card.Set)
	totalCopies := d.totalCopies(cardName)
	if totalCopies >= 2 {
		return AddCardResult{
			Card:        card,
			Added:       false,
			TotalCopies: totalCopies,
			SetCopies:   d.setCopies(cardName, cardSet),
		}, nil
	}

	for idx := range d.Cards {
		entry := &d.Cards[idx]
		if strings.EqualFold(entry.Name, cardName) && strings.EqualFold(entry.Set, cardSet) {
			if entry.Count >= 2 {
				return AddCardResult{
					Card:        card,
					Entry:       *entry,
					Added:       false,
					TotalCopies: totalCopies,
					SetCopies:   entry.Count,
				}, nil
			}
			entry.Count++
			return AddCardResult{
				Card:        card,
				Entry:       *entry,
				Added:       true,
				TotalCopies: totalCopies + 1,
				SetCopies:   entry.Count,
			}, nil
		}
	}

	entry := CardEntry{Name: cardName, Set: cardSet, Count: 1}
	d.Cards = append(d.Cards, entry)
	return AddCardResult{
		Card:        card,
		Entry:       entry,
		Added:       true,
		TotalCopies: totalCopies + 1,
		SetCopies:   1,
	}, nil
}

func (d *Deck) RemoveCard(index int) (CardEntry, error) {
	if index < 0 || index >= len(d.Cards) {
		return CardEntry{}, fmt.Errorf("index %d out of range", index)
	}

	entry := d.Cards[index]
	if entry.Count > 1 {
		d.Cards[index].Count--
		entry.Count = d.Cards[index].Count
		return entry, nil
	}

	d.Cards = append(d.Cards[:index], d.Cards[index+1:]...)
	return entry, nil
}

func (d *Deck) RecordBattle(result, opponent string, now time.Time) error {
	outcome := strings.ToUpper(strings.TrimSpace(result))
	if outcome != "W" && outcome != "L" {
		return fmt.Errorf("invalid outcome %q", result)
	}

	record := BattleRecord{
		Date:     now.Format("2006-01-02 15:04:05"),
		Result:   outcome,
		Opponent: strings.TrimSpace(opponent),
	}
	if record.Opponent == "" {
		record.Opponent = "Unknown"
	}

	d.BattleHistory = append(d.BattleHistory, record)
	return nil
}

func (d *Deck) Stats() Stats {
	stats := Stats{
		LossByOpponent: make(map[string]int),
	}
	stats.TotalBattles = len(d.BattleHistory)
	if stats.TotalBattles == 0 {
		return stats
	}

	for _, battle := range d.BattleHistory {
		switch {
		case strings.EqualFold(battle.Result, "W"):
			stats.Wins++
		case strings.EqualFold(battle.Result, "L"):
			stats.LossByOpponent[battle.Opponent]++
		}
	}
	stats.Losses = stats.TotalBattles - stats.Wins
	stats.WinPercentage = (float64(stats.Wins) / float64(stats.TotalBattles)) * 100
	return stats
}

func (d *Deck) totalCopies(cardName string) int {
	total := 0
	for _, entry := range d.Cards {
		if strings.EqualFold(entry.Name, cardName) {
			total += entry.Count
		}
	}
	return total
}

func (d *Deck) setCopies(cardName, cardSet string) int {
	for _, entry := range d.Cards {
		if strings.EqualFold(entry.Name, cardName) && strings.EqualFold(entry.Set, cardSet) {
			return entry.Count
		}
	}
	return 0
}

func parseRemoteCardNumber(number json.Number) (int, error) {
	value := strings.TrimSpace(number.String())
	if value == "" {
		return 0, fmt.Errorf("empty card number")
	}
	return strconv.Atoi(value)
}
