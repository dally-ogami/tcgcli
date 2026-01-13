package tcg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	cardsURL = "https://raw.githubusercontent.com/flibustier/pokemon-tcg-pocket-database/main/dist/cards.json"
	setsURL  = "https://raw.githubusercontent.com/flibustier/pokemon-tcg-pocket-database/main/dist/sets.json"
)

type remoteCard struct {
	Set    string            `json:"set"`
	Number json.Number       `json:"number"`
	Label  map[string]string `json:"label"`
}

type remoteSet struct {
	Code  string            `json:"code"`
	Label map[string]string `json:"label"`
}

func LoadValidCards() ([]Card, CardsSource, error, error) {
	cards, err := fetchRemoteCards()
	if err == nil {
		return cards, CardsSourceRemote, nil, nil
	}

	localCards, localErr := loadLocalCards()
	if localErr != nil {
		return nil, CardsSourceNone, nil, fmt.Errorf("remote error: %w; local error: %v", err, localErr)
	}

	return localCards, CardsSourceLocal, err, nil
}

func fetchRemoteCards() ([]Card, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	var rawCards []remoteCard
	if err := fetchJSON(client, cardsURL, &rawCards); err != nil {
		return nil, err
	}

	var rawSets []remoteSet
	if err := fetchJSON(client, setsURL, &rawSets); err != nil {
		return nil, err
	}

	setMap := make(map[string]string)
	for _, s := range rawSets {
		if s.Code == "" {
			continue
		}
		setMap[strings.ToLower(s.Code)] = pickLabel(s.Label, s.Code)
	}

	var cards []Card
	for _, raw := range rawCards {
		setCode := strings.TrimSpace(raw.Set)
		if setCode == "" {
			continue
		}

		number, err := parseRemoteCardNumber(raw.Number)
		if err != nil {
			continue
		}

		name := strings.TrimSpace(pickLabel(raw.Label, ""))
		if name == "" {
			continue
		}

		setName := setMap[strings.ToLower(setCode)]
		if setName == "" {
			setName = setCode
		}

		cards = append(cards, Card{
			Name: name,
			Set:  fmt.Sprintf("%s (%s)", setName, setCode),
			ID:   fmt.Sprintf("%s-%03d", strings.ToLower(setCode), number),
		})
	}

	return cards, nil
}

func fetchJSON(client *http.Client, url string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	return decoder.Decode(target)
}

func pickLabel(label map[string]string, fallback string) string {
	if label == nil {
		return fallback
	}
	if eng, ok := label["eng"]; ok && strings.TrimSpace(eng) != "" {
		return eng
	}
	if en, ok := label["en"]; ok && strings.TrimSpace(en) != "" {
		return en
	}
	return fallback
}

func loadLocalCards() ([]Card, error) {
	file, err := os.Open("valid_cards.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cards []Card
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cards); err != nil {
		return nil, err
	}

	for idx := range cards {
		cards[idx].Name = strings.TrimSpace(cards[idx].Name)
		cards[idx].Set = strings.TrimSpace(cards[idx].Set)
		cards[idx].ID = strings.TrimSpace(cards[idx].ID)
	}

	return cards, nil
}
