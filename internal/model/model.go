package model

import (
	"fmt"
	"time"
)

const (
	Lose SpinResult = iota
	Win
	SuperWin
)

type (
	SpinResult byte

	User struct {
		ID        string `json:"id,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Login     string `json:"login"`
		Password  string `json:"password"`
		Balance   int64  `json:"balance,omitempty"`
	}

	Spin struct {
		ID          string     `json:"id"`
		UserID      string     `json:"user_id"`
		Combination string     `json:"combination"`
		Result      SpinResult `json:"spin_result"`
		BetAmount   int64      `json:"bet_amount"`
		WinAmount   int64      `json:"win_amount"`
		CreatedAt   time.Time  `json:"created_at"`
	}
)

func (sr SpinResult) String() string {
	switch sr {
	case Lose:
		return "Lose"
	case Win:
		return "Win"
	case SuperWin:
		return "SuperWin"
	}

	panic(fmt.Sprintf("unknown spin result type: %d", sr))
}
