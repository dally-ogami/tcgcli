package tcg

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

type DeckLoadStatus string

type CardsSource string

type AddCardResult struct {
	Card        Card
	Entry       CardEntry
	Added       bool
	TotalCopies int
	SetCopies   int
}

const (
	DeckLoadNew    DeckLoadStatus = "new"
	DeckLoadLoaded DeckLoadStatus = "loaded"
	DeckLoadReset  DeckLoadStatus = "reset"
)

const (
	CardsSourceRemote CardsSource = "remote"
	CardsSourceLocal  CardsSource = "local"
	CardsSourceNone   CardsSource = "none"
)
