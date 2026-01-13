package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"tcgcli/tcg"
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

type DeckManager struct {
	DecksDir    string
	CurrentDeck *tcg.Deck
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
	return &DeckManager{DecksDir: decksDir}, nil
}

func (m *DeckManager) ListExistingDecks() ([]string, error) {
	manager, err := tcg.NewDeckManager(m.DecksDir)
	if err != nil {
		return nil, err
	}
	return manager.ListExistingDecks()
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

	manager, err := tcg.NewDeckManager(m.DecksDir)
	if err != nil {
		return err
	}
	deck, err := manager.CreateDeck(deckName)
	if errors.Is(err, os.ErrExist) {
		fmt.Printf("%sA deck with that name already exists.%s\n", colorRed, colorReset)
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("%sNew deck '%s' created.%s\n", colorGreen, deckName, colorReset)
	m.handleDeckLoadMessages(deck)
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
	manager, err := tcg.NewDeckManager(m.DecksDir)
	if err != nil {
		return err
	}
	deck, err := manager.LoadDeck(selectedDeck)
	if err != nil {
		return err
	}

	fmt.Printf("%sDeck '%s' loaded.%s\n", colorGreen, selectedDeck, colorReset)
	m.handleDeckLoadMessages(deck)
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

func mainMenu(reader *bufio.Reader, deck *tcg.Deck) {
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
			availableCards := deck.ListAvailableCards()
			if len(availableCards) == 0 {
				fmt.Printf("%sNo valid cards available.%s\n", colorRed, colorReset)
			} else {
				fmt.Printf("%s\nAvailable Cards:%s\n", colorCyan, colorReset)
				for _, card := range availableCards {
					fmt.Printf(" - %s (Set: %s, ID: %s)\n", formatForDisplay(card.Name), formatForDisplay(card.Set), card.ID)
				}
			}
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
				addCard(reader, deck, searchTerm)
			} else if next != "main" && next != "" {
				fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
			}
		case "1":
			searchTerm, err := prompt(reader, fmt.Sprintf("%s\nEnter search term (name or set): %s", colorMagenta, colorReset))
			if err != nil {
				continue
			}
			addCard(reader, deck, searchTerm)
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
					addCard(reader, deck, searchTerm)
				} else if contLower == "no" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "2":
			viewDeck(deck)
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
					removeCard(deck, index)
				}
			} else if action != "main" && action != "" {
				fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
			}
		case "3":
			if len(deck.Cards) == 0 {
				fmt.Printf("%sCannot remove from an empty deck.%s\n", colorRed, colorReset)
				continue
			}
			viewDeck(deck)
			index, ok := promptForIndex(reader, len(deck.Cards))
			if ok {
				removeCard(deck, index)
			}
			for len(deck.Cards) > 0 {
				cont, err := prompt(reader, fmt.Sprintf("%s\nDo you want to remove another card? (yes/no): %s", colorMagenta, colorReset))
				if err != nil {
					break
				}
				contLower := strings.ToLower(strings.TrimSpace(cont))
				if contLower == "yes" {
					viewDeck(deck)
					if len(deck.Cards) == 0 {
						fmt.Printf("%sCannot remove from an empty deck.%s\n", colorRed, colorReset)
						break
					}
					index, ok := promptForIndex(reader, len(deck.Cards))
					if ok {
						removeCard(deck, index)
					}
				} else if contLower == "no" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "4":
			recordBattle(reader, deck)
			for {
				next, err := prompt(reader, fmt.Sprintf("%s\nDo you want to record another battle or go back to the main menu? (add/main): %s", colorMagenta, colorReset))
				if err != nil {
					break
				}
				nextLower := strings.ToLower(strings.TrimSpace(next))
				if nextLower == "add" {
					recordBattle(reader, deck)
				} else if nextLower == "main" {
					break
				} else {
					fmt.Printf("%sInvalid choice. Going back to the main menu.%s\n", colorRed, colorReset)
					break
				}
			}
		case "5":
			showStatistics(deck)
		case "6":
			if err := deck.Save(); err != nil {
				fmt.Printf("%sFailed to save deck: %v%s\n", colorRed, err, colorReset)
			} else {
				fmt.Printf("%sDeck '%s' saved successfully!%s\n", colorGreen, deck.Name, colorReset)
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

func formatForDisplay(value string) string {
	return strings.TrimSpace(value)
}

func addCard(reader *bufio.Reader, deck *tcg.Deck, searchTerm string) {
	results := deck.SearchCards(searchTerm)
	if len(results) == 0 {
		fmt.Printf("%sNo valid card found matching '%s'.%s\n", colorRed, searchTerm, colorReset)
		return
	}

	var selected tcg.Card
	if len(results) > 1 {
		fmt.Printf("%s\nMultiple matches found:%s\n", colorCyan, colorReset)
		for idx, card := range results {
			fmt.Printf("  %d. %s (Set: %s, ID: %s)\n", idx+1, formatForDisplay(card.Name), formatForDisplay(card.Set), card.ID)
		}
		choiceStr, err := prompt(reader, fmt.Sprintf("%sEnter the number of the card you want to add: %s", colorWhite, colorReset))
		if err != nil {
			return
		}
		choice, err := strconv.Atoi(choiceStr)
		if err != nil || choice < 1 || choice > len(results) {
			fmt.Printf("%sInvalid selection.%s\n", colorRed, colorReset)
			return
		}
		selected = results[choice-1]
	} else {
		selected = results[0]
	}

	result, err := deck.AddCardByID(selected.ID)
	if err != nil {
		fmt.Printf("%sFailed to add card: %v%s\n", colorRed, err, colorReset)
		return
	}

	cardName := formatForDisplay(result.Card.Name)
	cardSet := formatForDisplay(result.Card.Set)

	if !result.Added {
		if result.TotalCopies >= 2 {
			fmt.Printf("%sWarning: Already have 2 copies of %s (across all sets). Cannot add more.%s\n", colorYellow, cardName, colorReset)
			return
		}
		fmt.Printf("%sWarning: Already have 2 copies of %s from %s.%s\n", colorYellow, cardName, cardSet, colorReset)
		return
	}

	if result.SetCopies == 2 {
		fmt.Printf("%s%s from %s added. You now have 2 copies in this set.%s\n", colorGreen, cardName, cardSet, colorReset)
	} else if result.SetCopies == 1 && result.TotalCopies == 1 {
		fmt.Printf("%s%s from %s added to your deck.%s\n", colorGreen, cardName, cardSet, colorReset)
	} else {
		fmt.Printf("%s%s from %s added.%s\n", colorGreen, cardName, cardSet, colorReset)
	}
}

func viewDeck(deck *tcg.Deck) {
	if len(deck.Cards) == 0 {
		fmt.Printf("%sYour deck is empty.%s\n", colorYellow, colorReset)
		return
	}

	fmt.Printf("%s\nDeck: %s%s\n", colorLightCyan, deck.Name, colorReset)
	for idx, entry := range deck.Cards {
		fmt.Printf("%s  %d. %s x %d from %s%s\n", colorLightCyan, idx+1, entry.Name, entry.Count, entry.Set, colorReset)
	}
}

func removeCard(deck *tcg.Deck, index int) {
	entry, err := deck.RemoveCard(index)
	if err != nil {
		fmt.Printf("%sInvalid index. Nothing was removed.%s\n", colorRed, colorReset)
		return
	}
	if entry.Count > 1 {
		fmt.Printf("%sOne copy of %s from %s removed. Now you have %d copy(ies).%s\n", colorGreen, entry.Name, entry.Set, entry.Count, colorReset)
		return
	}
	fmt.Printf("%s%s from %s removed from your deck.%s\n", colorGreen, entry.Name, entry.Set, colorReset)
}

func recordBattle(reader *bufio.Reader, deck *tcg.Deck) {
	outcome, err := prompt(reader, fmt.Sprintf("%sEnter battle outcome (W for win, L for loss): %s", colorWhite, colorReset))
	if err != nil {
		return
	}
	opponent, err := prompt(reader, fmt.Sprintf("%sEnter opponent deck details (or other metadata): %s", colorWhite, colorReset))
	if err != nil {
		return
	}
	if err := deck.RecordBattle(outcome, opponent, time.Now()); err != nil {
		fmt.Printf("%sInvalid outcome. Use 'W' or 'L'.%s\n", colorRed, colorReset)
		return
	}
	fmt.Printf("%sBattle record added for deck '%s'.%s\n", colorGreen, deck.Name, colorReset)
}

func showStatistics(deck *tcg.Deck) {
	stats := deck.Stats()
	if stats.TotalBattles == 0 {
		fmt.Printf("%sNo battle records to show statistics.%s\n", colorYellow, colorReset)
		return
	}

	fmt.Printf("%s\nBattle Statistics for '%s':%s\n", colorCyan, deck.Name, colorReset)
	fmt.Printf("%s  Total Battles: %d%s\n", colorCyan, stats.TotalBattles, colorReset)
	fmt.Printf("%s  Wins: %d%s\n", colorCyan, stats.Wins, colorReset)
	fmt.Printf("%s  Losses: %d%s\n", colorCyan, stats.Losses, colorReset)
	fmt.Printf("%s  Win Percentage: %.2f%%%s\n", colorCyan, stats.WinPercentage, colorReset)

	fmt.Printf("%s\nWin/Loss Graph:%s\n", colorBlue, colorReset)
	fmt.Printf("%sWins  : %s%s\n", colorGreen, strings.Repeat("*", stats.Wins), colorReset)
	fmt.Printf("%sLosses: %s%s\n", colorRed, strings.Repeat("*", stats.Losses), colorReset)

	if len(stats.LossByOpponent) > 0 {
		fmt.Printf("%s\nLoss Frequency by Opponent Deck:%s\n", colorLightMagenta, colorReset)
		for opponent, count := range stats.LossByOpponent {
			fmt.Printf("%s  %s: %d loss(es)%s\n", colorLightMagenta, opponent, count, colorReset)
		}
	}
}

func (m *DeckManager) handleDeckLoadMessages(deck *tcg.Deck) {
	switch deck.CardsSource {
	case tcg.CardsSourceRemote:
		fmt.Printf("%sLoaded latest card data from online database.%s\n", colorGreen, colorReset)
	case tcg.CardsSourceLocal:
		if deck.CardsLoadError != nil {
			fmt.Printf("%sWarning: Could not fetch latest card data (%v). Using local cache.%s\n", colorYellow, deck.CardsLoadError, colorReset)
		}
	}

	switch deck.LoadStatus {
	case tcg.DeckLoadNew:
		fmt.Printf("%sDeck file '%s' not found. Starting new deck '%s'.%s\n", colorYellow, deck.FilePath, deck.Name, colorReset)
	case tcg.DeckLoadReset:
		fmt.Printf("%sError decoding %s. Starting with an empty deck.%s\n", colorRed, deck.FilePath, colorReset)
	case tcg.DeckLoadLoaded:
		fmt.Printf("%sDeck '%s' loaded from %s.%s\n", colorGreen, deck.Name, deck.FilePath, colorReset)
	}
}
