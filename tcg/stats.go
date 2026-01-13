package tcg

type Stats struct {
	TotalBattles   int
	Wins           int
	Losses         int
	WinPercentage  float64
	LossByOpponent map[string]int
}
