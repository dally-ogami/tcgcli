package tcg

type Stats struct {
	TotalBattles   int            `json:"totalBattles"`
	Wins           int            `json:"wins"`
	Losses         int            `json:"losses"`
	WinPercentage  float64        `json:"winPercentage"`
	LossByOpponent map[string]int `json:"lossByOpponent"`
}
