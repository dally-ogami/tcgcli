package main

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"tcgcli/tcg"
)

const defaultAddr = ":8080"

//go:embed web/index.html web/assets/*
var webFS embed.FS

type server struct {
	decksDir string

	cardsOnce sync.Once
	cards     []tcg.Card
	cardsFrom tcg.CardsSource
	cardsWarn error
	cardsErr  error
}

type errorResponse struct {
	Error string `json:"error"`
}

type createDeckRequest struct {
	Name string `json:"name"`
}

type addCardRequest struct {
	CardID string `json:"card_id"`
}

type recordBattleRequest struct {
	Result   string `json:"result"`
	Opponent string `json:"opponent"`
}

type deckResponse struct {
	Name         string             `json:"name"`
	Cards        []tcg.CardEntry    `json:"cards"`
	Battles      []tcg.BattleRecord `json:"battles"`
	Stats        tcg.Stats          `json:"stats"`
	LoadStatus   tcg.DeckLoadStatus `json:"load_status"`
	CardsSource  tcg.CardsSource    `json:"cards_source"`
	CardsWarning string             `json:"cards_warning,omitempty"`
}

func main() {
	addr := defaultAddr
	if value := strings.TrimSpace(os.Getenv("TCG_WEB_ADDR")); value != "" {
		addr = value
	}

	srv := &server{decksDir: "decks"}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", srv.handleAPI)
	mux.HandleFunc("/", srv.handleIndex)

	assets, err := fs.Sub(webFS, "web/assets")
	if err != nil {
		log.Fatalf("failed to load assets: %v", err)
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	log.Printf("TCG web server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	index, err := webFS.ReadFile("web/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load UI")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index)
}

func (s *server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	path = strings.Trim(path, "/")

	if path == "decks" {
		s.handleDecks(w, r)
		return
	}

	if path == "cards" {
		s.handleCards(w, r)
		return
	}

	if strings.HasPrefix(path, "decks/") {
		s.handleDeck(w, r, strings.TrimPrefix(path, "decks/"))
		return
	}

	writeError(w, http.StatusNotFound, "unknown endpoint")
}

func (s *server) handleDecks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		manager, err := tcg.NewDeckManager(s.decksDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		decks, err := manager.ListExistingDecks()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"decks": decks})
	case http.MethodPost:
		var req createDeckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "deck name is required")
			return
		}

		manager, err := tcg.NewDeckManager(s.decksDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		deck, err := manager.CreateDeck(name)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				writeError(w, http.StatusConflict, "deck already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := deck.Save(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, s.toDeckResponse(deck))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleDeck(w http.ResponseWriter, r *http.Request, deckPath string) {
	segments := strings.Split(deckPath, "/")
	if len(segments) == 0 {
		writeError(w, http.StatusNotFound, "missing deck name")
		return
	}
	deckName, err := url.PathUnescape(segments[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid deck name")
		return
	}
	deckName = strings.TrimSpace(deckName)
	if deckName == "" {
		writeError(w, http.StatusBadRequest, "invalid deck name")
		return
	}

	if len(segments) == 1 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		deck, err := s.loadDeck(deckName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, s.toDeckResponse(deck))
		return
	}

	switch segments[1] {
	case "cards":
		s.handleDeckCards(w, r, deckName, segments[2:])
	case "battles":
		s.handleDeckBattles(w, r, deckName)
	default:
		writeError(w, http.StatusNotFound, "unknown deck endpoint")
	}
}

func (s *server) handleDeckCards(w http.ResponseWriter, r *http.Request, deckName string, segments []string) {
	switch r.Method {
	case http.MethodPost:
		if len(segments) != 0 {
			writeError(w, http.StatusNotFound, "unknown cards endpoint")
			return
		}
		var req addCardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		cardID := strings.TrimSpace(req.CardID)
		if cardID == "" {
			writeError(w, http.StatusBadRequest, "card ID is required")
			return
		}

		deck, err := s.loadDeck(deckName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if _, err := deck.AddCardByID(cardID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := deck.Save(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, s.toDeckResponse(deck))
	case http.MethodDelete:
		if len(segments) != 1 {
			writeError(w, http.StatusBadRequest, "card index required")
			return
		}
		index, err := strconv.Atoi(segments[0])
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid card index")
			return
		}
		deck, err := s.loadDeck(deckName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if _, err := deck.RemoveCard(index); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := deck.Save(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, s.toDeckResponse(deck))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handleDeckBattles(w http.ResponseWriter, r *http.Request, deckName string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req recordBattleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	result := strings.TrimSpace(req.Result)
	if result == "" {
		writeError(w, http.StatusBadRequest, "result is required")
		return
	}

	deck, err := s.loadDeck(deckName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := deck.RecordBattle(result, strings.TrimSpace(req.Opponent), time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := deck.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.toDeckResponse(deck))
}

func (s *server) handleCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("search"))
	cards, _, _, err := s.loadCards()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	matches := filterCards(cards, query)
	writeJSON(w, http.StatusOK, map[string]any{"cards": matches})
}

func (s *server) loadCards() ([]tcg.Card, tcg.CardsSource, error, error) {
	s.cardsOnce.Do(func() {
		cards, source, warn, err := tcg.LoadValidCards()
		s.cards = cards
		s.cardsFrom = source
		s.cardsWarn = warn
		s.cardsErr = err
	})
	return s.cards, s.cardsFrom, s.cardsWarn, s.cardsErr
}

func (s *server) loadDeck(name string) (*tcg.Deck, error) {
	manager, err := tcg.NewDeckManager(s.decksDir)
	if err != nil {
		return nil, err
	}
	return manager.LoadDeck(name)
}

func (s *server) toDeckResponse(deck *tcg.Deck) deckResponse {
	warning := ""
	if deck.CardsLoadError != nil {
		warning = deck.CardsLoadError.Error()
	}

	return deckResponse{
		Name:         deck.Name,
		Cards:        deck.Cards,
		Battles:      deck.BattleHistory,
		Stats:        deck.Stats(),
		LoadStatus:   deck.LoadStatus,
		CardsSource:  deck.CardsSource,
		CardsWarning: warning,
	}
}

func filterCards(cards []tcg.Card, term string) []tcg.Card {
	if term == "" {
		if len(cards) > 200 {
			return cards[:200]
		}
		return cards
	}

	normalized := strings.ToLower(term)
	var matches []tcg.Card
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card.Name), normalized) || strings.Contains(strings.ToLower(card.Set), normalized) || strings.Contains(strings.ToLower(card.ID), normalized) {
			matches = append(matches, card)
		}
	}
	return matches
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
