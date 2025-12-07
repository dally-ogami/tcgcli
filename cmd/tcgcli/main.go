package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	colorReset        = "\033[0m"
	colorGreen        = "\033[32m"
	colorYellow       = "\033[33m"
	colorRed          = "\033[31m"
	colorCyan         = "\033[36m"
	colorLightCyan    = "\033[96m"
	colorMagenta      = "\033[35m"
	colorLightMagenta = "\033[95m"
	colorBlue         = "\033[34m"
	colorWhite        = "\033[37m"
)

const (
	cardsURL = "https://raw.githubusercontent.com/flibustier/pokemon-tcg-pocket-database/main/dist/cards.json"
	setsURL  = "https://raw.githubusercontent.com/flibustier/pokemon-tcg-pocket-database/main/dist/sets.json"
)

type Card struct {
	Name string `json:"name"`
	Set  string `json:"set"`
	ID   string `json:"id"`
}

type CardEntry struct {
	Name  string `json:"name"`
	Set   string `json:"set"`
	Count int    `json:"count"`
}

type BattleRecord struct {
	Date     string `json:"date"`
	Result   string `json:"result"`
	Opponent string `json:"opponent"`
}

type deckFileData struct {
	Cards         []CardEntry    `json:"cards"`
	BattleHistory []BattleRecord `json:"battle_history"`
}

type Deck struct {
	Name          string
	FilePath      string
	Cards         []CardEntry
	BattleHistory []BattleRecord
	ValidCards    []Card
}

type DeckManager struct {
	DecksDir    string
	CurrentDeck *Deck
}

type remoteCard struct {
	Set    string            `json:"set"`
	Number json.Number       `json:"number"`
	Label  map[string]string `json:"label"`
}

type remoteSet struct {
	Code  string            `json:"code"`
	Label map[string]string `json:"label"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	manager, err := NewDeckManager("decks")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sFailed to initialize deck manager: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if err := manager.SelectDeck(reader); err != nil {
		fmt.Fprintf(os.Stderr, "%sError selecting deck: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if manager.CurrentDeck != nil {
		mainMenu(reader, manager.CurrentDeck)
	}
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

func (m *DeckManager) CreateNewDeck(reader *bufio.Reader) error {
	deckName, err := prompt(reader, fmt.Sprintf("%sEnter a name for your new deck: %s", colorWhite, colorReset))
	if err != nil {
		return err
	}
	if deckName == "" {
		fmt.Printf("%sDeck name cannot be empty.%s\n", colorRed, colorReset)
		return nil
	}

	deckFile := filepath.Join(m.DecksDir, deckName+".json")
	if _, err := os.Stat(deckFile); err == nil {
		fmt.Printf("%sA deck with that name already exists.%s\n", colorRed, colorReset)
		return nil
	}

	deck, err := NewDeck(deckName, deckFile)
	if err != nil {
		return err
	}

	fmt.Printf("%sNew deck '%s' created.%s\n", colorGreen, deckName, colorReset)
	m.CurrentDeck = deck
	return nil
}

func (m *DeckManager) LoadExistingDeck(reader *bufio.Reader) error {
	decks, err := m.ListExistingDecks()
	if err != nil {
		return err
	}
	if len(decks) == 0 {
		fmt.Printf("%sNo saved decks found.%s\n", colorYellow, colorReset)
		return nil
	}

	fmt.Printf("%s\nExisting decks:%s\n", colorCyan, colorReset)
	for idx, deckName := range decks {
		fmt.Printf("%s  %d. %s%s\n", colorCyan, idx+1, deckName, colorReset)
	}

	choiceStr, err := prompt(reader, fmt.Sprintf("%sEnter the number of the deck to load: %s", colorWhite, colorReset))
	if err != nil {
		return err
	}
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(decks) {
		fmt.Printf("%sInvalid selection.%s\n", colorRed, colorReset)
		return nil
	}

	selectedDeck := decks[choice-1]
	deckFile := filepath.Join(m.DecksDir, selectedDeck+".json")
	deck, err := NewDeck(selectedDeck, deckFile)
	if err != nil {
		return err
	}

	fmt.Printf("%sDeck '%s' loaded.%s\n", colorGreen, selectedDeck, colorReset)
	m.CurrentDeck = deck
	return nil
}

func (m *DeckManager) SelectDeck(reader *bufio.Reader) error {
	for {
		fmt.Printf("%s\nDeck Manager Options:%s\n", colorMagenta, colorReset)
		fmt.Println("  1: Create a new deck")
		fmt.Println("  2: Load an existing deck")
		fmt.Println("  3: Exit")

		choice, err := prompt(reader, fmt.Sprintf("%sEnter your choice (1-3): %s", colorWhite, colorReset))
		if err != nil {
			return err
		}

		switch choice {
		case "1":
			if err := m.CreateNewDeck(reader); err != nil {
				return err
			}
			if m.CurrentDeck != nil {
				return nil
			}
		case "2":
			if err := m.LoadExistingDeck(reader); err != nil {
				return err
			}
			if m.CurrentDeck != nil {
				return nil
			}
		case "3":
			fmt.Printf("%sGoodbye!%s\n", colorGreen, colorReset)
			return nil
		default:
			fmt.Printf("%sInvalid option. Please try again.%s\n", colorRed, colorReset)
		}
	}
}

func NewDeck(name, filePath string) (*Deck, error) {
	deck := &Deck{
		Name:     name,
		FilePath: filePath,
	}

	deck.ValidCards = loadValidCards()
	if err := deck.loadDeckFile(); err != nil {
		return nil, err
	}

	return deck, nil
}

func (d *Deck) loadDeckFile() error {
	if _, err := os.Stat(d.FilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%sDeck file '%s' not found. Starting new deck '%s'.%s\n", colorYellow, d.FilePath, d.Name, colorReset)
		d.Cards = []CardEntry{}
		d.BattleHistory = []BattleRecord{}
		return nil
	} else if err != nil {
		return err
	}

	file, err := os.Open(d.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data deckFileData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		fmt.Printf("%sError decoding %s. Starting with an empty deck.%s\n", colorRed, d.FilePath, colorReset)
		d.Cards = []CardEntry{}
		d.BattleHistory = []BattleRecord{}
		return nil
	}

	d.Cards = data.Cards
	d.BattleHistory = data.BattleHistory
	fmt.Printf("%sDeck '%s' loaded from %s.%s\n", colorGreen, d.Name, d.FilePath, colorReset)
	return nil
}

func (d *Deck) SaveDeck() error {
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
	if err := encoder.Encode(&data); err != nil {
		return err
	}

	fmt.Printf("%sDeck '%s' saved successfully!%s\n", colorGreen, d.Name, colorReset)
	return nil
}

func (d *Deck) ListAvailableCards() {
	if len(d.ValidCards) == 0 {
		fmt.Printf("%sNo valid cards available.%s\n", colorRed, colorReset)
		return
	}

	fmt.Printf("%s\nAvailable Cards:%s\n", colorCyan, colorReset)
	for _, card := range d.ValidCards {
		fmt.Printf(" - %s (Set: %s, ID: %s)\n", formatForDisplay(card.Name), formatForDisplay(card.Set), card.ID)
	}
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

func (d *Deck) totalCopies(cardName string) int {
	total := 0
	for _, entry := range d.Cards {
		if strings.EqualFold(entry.Name, cardName) {
			total += entry.Count
		}
	}
	return total
}

func (d *Deck) AddCard(reader *bufio.Reader, searchTerm string) bool {
	results := d.SearchCards(searchTerm)
	if len(results) == 0 {
		fmt.Printf("%sNo valid card found matching '%s'.%s\n", colorRed, searchTerm, colorReset)
		return false
	}

	var selected Card
	if len(results) > 1 {
		fmt.Printf("%s\nMultiple matches found:%s\n", colorCyan, colorReset)
		for idx, card := range results {
			fmt.Printf("  %d. %s (Set: %s, ID: %s)\n", idx+1, formatForDisplay(card.Name), formatForDisplay(card.Set), card.ID)
		}
		choiceStr, err := prompt(reader, fmt.Sprintf("%sEnter the number of the card you want to add: %s", colorWhite, colorReset))
		if err != nil {
			return false
		}
		choice, err := strconv.Atoi(choiceStr)
		if err != nil || choice < 1 || choice > len(results) {
			fmt.Printf("%sInvalid selection.%s\n", colorRed, colorReset)
			return false
		}
		selected = results[choice-1]
	} else {
		selected = results[0]
	}

	cardName := formatForDisplay(selected.Name)
	cardSet := formatForDisplay(selected.Set)

	if d.totalCopies(cardName) >= 2 {
		fmt.Printf("%sWarning: Already have 2 copies of %s (across all sets). Cannot add more.%s\n", colorYellow, cardName, colorReset)
		return false
	}

	for idx := range d.Cards {
		entry := &d.Cards[idx]
		if strings.EqualFold(entry.Name, cardName) && strings.EqualFold(entry.Set, cardSet) {
			if entry.Count >= 2 {
				fmt.Printf("%sWarning: Already have 2 copies of %s from %s.%s\n", colorYellow, cardName, cardSet, colorReset)
				return false
			}
			entry.Count++
			if entry.Count == 2 {
				fmt.Printf("%s%s from %s added. You now have 2 copies in this set.%s\n", colorGreen, cardName, cardSet, colorReset)
			} else {
				fmt.Printf("%s%s from %s added.%s\n", colorGreen, cardName, cardSet, colorReset)
			}
			return true
		}
	}

	d.Cards = append(d.Cards, CardEntry{Name: cardName, Set: cardSet, Count: 1})
	fmt.Printf("%s%s from %s added to your deck.%s\n", colorGreen, cardName, cardSet, colorReset)
	return true
}

func (d *Deck) ViewDeck() {
	if len(d.Cards) == 0 {
		fmt.Printf("%sYour deck is empty.%s\n", colorYellow, colorReset)
		return
	}

	fmt.Printf("%s\nDeck: %s%s\n", colorLightCyan, d.Name, colorReset)
	for idx, entry := range d.Cards {
		fmt.Printf("%s  %d. %s x %d from %s%s\n", colorLightCyan, idx+1, entry.Name, entry.Count, entry.Set, colorReset)
	}
}

func (d *Deck) RemoveCard(index int) bool {
	if index < 0 || index >= len(d.Cards) {
		fmt.Printf("%sInvalid index. Nothing was removed.%s\n", colorRed, colorReset)
		return false
	}

	entry := &d.Cards[index]
	if entry.Count > 1 {
		entry.Count--
		fmt.Printf("%sOne copy of %s from %s removed. Now you have %d copy(ies).%s\n", colorGreen, entry.Name, entry.Set, entry.Count, colorReset)
		return true
	}

	removed := d.Cards[index]
	d.Cards = append(d.Cards[:index], d.Cards[index+1:]...)
	fmt.Printf("%s%s from %s removed from your deck.%s\n", colorGreen, removed.Name, removed.Set, colorReset)
	return true
}

func (d *Deck) RecordBattle(reader *bufio.Reader) {
	outcome, err := prompt(reader, fmt.Sprintf("%sEnter battle outcome (W for win, L for loss): %s", colorWhite, colorReset))
	if err != nil {
		return
	}
	outcome = strings.ToUpper(strings.TrimSpace(outcome))
	if outcome != "W" && outcome != "L" {
		fmt.Printf("%sInvalid outcome. Use 'W' or 'L'.%s\n", colorRed, colorReset)
		return
	}

	opponent, err := prompt(reader, fmt.Sprintf("%sEnter opponent deck details (or other metadata): %s", colorWhite, colorReset))
	if err != nil {
		return
	}

	record := BattleRecord{
		Date:     time.Now().Format("2006-01-02 15:04:05"),
		Result:   outcome,
		Opponent: opponent,
	}
	d.BattleHistory = append(d.BattleHistory, record)
	fmt.Printf("%sBattle record added for deck '%s'.%s\n", colorGreen, d.Name, colorReset)
}

func (d *Deck) ShowStatistics() {
	totalBattles := len(d.BattleHistory)
	if totalBattles == 0 {
		fmt.Printf("%sNo battle records to show statistics.%s\n", colorYellow, colorReset)
		return
	}

	wins := 0
	loseMatchups := make(map[string]int)
	for _, battle := range d.BattleHistory {
		if strings.EqualFold(battle.Result, "W") {
			wins++
		} else {
			loseMatchups[battle.Opponent]++
		}
	}
	losses := totalBattles - wins
	winPercentage := (float64(wins) / float64(totalBattles)) * 100

	fmt.Printf("%s\nBattle Statistics for '%s':%s\n", colorCyan, d.Name, colorReset)
	fmt.Printf("%s  Total Battles: %d%s\n", colorCyan, totalBattles, colorReset)
	fmt.Printf("%s  Wins: %d%s\n", colorCyan, wins, colorReset)
	fmt.Printf("%s  Losses: %d%s\n", colorCyan, losses, colorReset)
	fmt.Printf("%s  Win Percentage: %.2f%%%s\n", colorCyan, winPercentage, colorReset)

	fmt.Printf("%s\nWin/Loss Graph:%s\n", colorBlue, colorReset)
	fmt.Printf("%sWins  : %s%s\n", colorGreen, strings.Repeat("*", wins), colorReset)
	fmt.Printf("%sLosses: %s%s\n", colorRed, strings.Repeat("*", losses), colorReset)

	if len(loseMatchups) > 0 {
		fmt.Printf("%s\nLoss Frequency by Opponent Deck:%s\n", colorLightMagenta, colorReset)
		for opponent, count := range loseMatchups {
			fmt.Printf("%s  %s: %d loss(es)%s\n", colorLightMagenta, opponent, count, colorReset)
		}
	}
}

func mainMenu(reader *bufio.Reader, deck *Deck) {
	for {
		fmt.Printf("%s\nMain Menu:%s\n", colorMagenta, colorReset)
		fmt.Println("  0: List all available cards")
		fmt.Println("  1: Add a card to your deck (search by name or set)")
		fmt.Println("  2: View your deck")
		fmt.Println("  3: Remove a card from your deck")
		fmt.Println("  4: Record a battle outcome")
		fmt.Println("  5: Show deck battle statistics")
		fmt.Println("  6: Save and exit")

		choice, err := prompt(reader, fmt.Sprintf("%sEnter your choice (0-6): %s", colorWhite, colorReset))
		if err != nil {
			fmt.Printf("%sError reading input: %v%s\n", colorRed, err, colorReset)
			continue
		}

		switch choice {
		case "0":
			deck.ListAvailableCards()
			next, err := prompt(reader, fmt.Sprintf("%s\nDo you want to add a card or go back to the main menu? (add/main): %s", colorMagenta, colorReset))
			if err != nil {
				continue
			}
			next = strings.ToLower(strings.TrimSpace(next))
			if next == "add" {
				searchTerm, err := prompt(reader, fmt.Sprintf("%s\nEnter search term (name or set): %s", colorMagenta, colorReset))
				if err != nil {
					continue
				}
				deck.AddCard(reader, searchTerm)
			} else if next != "main" && next != "" {
				fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
			}
		case "1":
			searchTerm, err := prompt(reader, fmt.Sprintf("%s\nEnter search term (name or set): %s", colorMagenta, colorReset))
			if err != nil {
				continue
			}
			deck.AddCard(reader, searchTerm)
			for {
				cont, err := prompt(reader, fmt.Sprintf("%s\nDo you want to add another card? (yes/no): %s", colorMagenta, colorReset))
				if err != nil {
					break
				}
				contLower := strings.ToLower(strings.TrimSpace(cont))
				if contLower == "yes" {
					searchTerm, err := prompt(reader, fmt.Sprintf("%s\nEnter search term (name or set): %s", colorMagenta, colorReset))
					if err != nil {
						break
					}
					deck.AddCard(reader, searchTerm)
				} else if contLower == "no" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "2":
			deck.ViewDeck()
			action, err := prompt(reader, fmt.Sprintf("%s\nDo you want to remove a card or go back to the main menu? (rm/main): %s", colorMagenta, colorReset))
			if err != nil {
				continue
			}
			action = strings.ToLower(strings.TrimSpace(action))
			if action == "rm" {
				if len(deck.Cards) == 0 {
					fmt.Printf("%sCannot remove from an empty deck.%s\n", colorRed, colorReset)
					continue
				}
				index, ok := promptForIndex(reader, len(deck.Cards))
				if ok {
					deck.RemoveCard(index)
				}
			} else if action != "main" && action != "" {
				fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
			}
		case "3":
			if len(deck.Cards) == 0 {
				fmt.Printf("%sCannot remove from an empty deck.%s\n", colorRed, colorReset)
				continue
			}
			deck.ViewDeck()
			index, ok := promptForIndex(reader, len(deck.Cards))
			if ok {
				deck.RemoveCard(index)
			}
			for len(deck.Cards) > 0 {
				cont, err := prompt(reader, fmt.Sprintf("%s\nDo you want to remove another card? (yes/no): %s", colorMagenta, colorReset))
				if err != nil {
					break
				}
				contLower := strings.ToLower(strings.TrimSpace(cont))
				if contLower == "yes" {
					deck.ViewDeck()
					if len(deck.Cards) == 0 {
						fmt.Printf("%sCannot remove from an empty deck.%s\n", colorRed, colorReset)
						break
					}
					index, ok := promptForIndex(reader, len(deck.Cards))
					if ok {
						deck.RemoveCard(index)
					}
				} else if contLower == "no" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "4":
			deck.RecordBattle(reader)
			for {
				next, err := prompt(reader, fmt.Sprintf("%s\nDo you want to record another battle or go back to the main menu? (add/main): %s", colorMagenta, colorReset))
				if err != nil {
					break
				}
				nextLower := strings.ToLower(strings.TrimSpace(next))
				if nextLower == "add" {
					deck.RecordBattle(reader)
				} else if nextLower == "main" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "5":
			deck.ShowStatistics()
		case "6":
			if err := deck.SaveDeck(); err != nil {
				fmt.Printf("%sFailed to save deck: %v%s\n", colorRed, err, colorReset)
			} else {
				fmt.Printf("%s\nExiting. Your deck has been saved!%s\n", colorGreen, colorReset)
			}
			return
		default:
			fmt.Printf("%sInvalid choice. Please try again.%s\n", colorRed, colorReset)
		}
	}
}

func prompt(reader *bufio.Reader, message string) (string, error) {
	fmt.Print(message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptForIndex(reader *bufio.Reader, length int) (int, bool) {
	response, err := prompt(reader, fmt.Sprintf("%s\nEnter the position (number) of the card to remove: %s", colorMagenta, colorReset))
	if err != nil {
		return 0, false
	}
	value, err := strconv.Atoi(response)
	if err != nil {
		fmt.Printf("%sPlease enter a valid number.%s\n", colorRed, colorReset)
		return 0, false
	}
	index := value - 1
	if index < 0 || index >= length {
		fmt.Printf("%sInvalid index. Nothing was removed.%s\n", colorRed, colorReset)
		return 0, false
	}
	return index, true
}

func loadValidCards() []Card {
	cards, err := fetchRemoteCards()
	if err == nil {
		fmt.Printf("%sLoaded latest card data from online database.%s\n", colorGreen, colorReset)
		return cards
	}

	fmt.Printf("%sWarning: Could not fetch latest card data (%v). Using local cache.%s\n", colorYellow, err, colorReset)
	localCards, localErr := loadLocalCards()
	if localErr != nil {
		fmt.Printf("%sError: valid_cards.json not found or invalid (%v).%s\n", colorRed, localErr, colorReset)
		return nil
	}
	return localCards
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

		numberStr := strings.TrimSpace(raw.Number.String())
		if numberStr == "" {
			continue
		}
		number, err := strconv.Atoi(numberStr)
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

func formatForDisplay(value string) string {
	return strings.TrimSpace(value)
}
