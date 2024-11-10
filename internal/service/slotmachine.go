package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/apolunin/slotgame/internal/model"
	"math/big"
	"strings"
)

type (
	Combination [3]byte
	GenerateFn  func() (Combination, error)

	SlotMachine struct {
		generate GenerateFn
	}
)

var ErrInvalidBetAmount = errors.New("invalid bet amount")

func NewSlotMachine() *SlotMachine {
	return &SlotMachine{
		generate: randomCombination,
	}
}

func NewSlotMachineWithGenerateFn(genFn GenerateFn) *SlotMachine {
	return &SlotMachine{
		generate: genFn,
	}
}

func (c Combination) Type() model.SpinResult {
	set := make(map[byte]struct{}, len(c))

	for _, n := range c {
		set[n] = struct{}{}
	}

	switch len(set) {
	case 1:
		return model.SuperWin
	case 2:
		return model.Win
	}

	return model.Lose
}

func (c Combination) String() string {
	var str [len(c)]string

	for i := 0; i < len(c); i++ {
		str[i] = fmt.Sprint(c[i])
	}

	return strings.Join(str[:], ",")
}

func (sm *SlotMachine) Spin(betAmount int64) (Combination, int64, error) {
	if betAmount <= 0 {
		return Combination{0, 0, 0}, 0, fmt.Errorf(
			"%w: bet amount must be positive",
			ErrInvalidBetAmount,
		)
	}

	c, err := sm.generate()

	if err != nil {
		return c, 0, fmt.Errorf("failed to generate combination: %w", err)
	}

	var winAmount int64

	switch c.Type() {
	case model.Lose:
		winAmount = -betAmount
	case model.Win:
		winAmount = 2 * betAmount
	case model.SuperWin:
		winAmount = 10 * betAmount
	}

	return c, winAmount, nil
}

func randomCombination() (Combination, error) {
	var res Combination

	for i := 0; i < len(res); i++ {
		r, err := randomNum()

		if err != nil {
			return res, fmt.Errorf("failed to generate random number: %w", err)
		}

		res[i] = r
	}

	return res, nil
}

func randomNum() (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9))

	if err != nil {
		return 0, err
	}

	return byte(n.Int64() + 1), nil
}
