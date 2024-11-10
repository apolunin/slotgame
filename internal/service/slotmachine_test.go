package service_test

import (
	"errors"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/service"
	"testing"
)

func TestSlotMachine_Spin(t *testing.T) {
	type (
		expected struct {
			err error
			res model.SpinResult
			win int64
		}

		test struct {
			name     string
			comb     service.Combination
			bet      int64
			err      error
			expected expected
		}
	)

	tests := []test{
		{
			name: "success - super win",
			comb: service.Combination{7, 7, 7},
			bet:  100,
			err:  nil,
			expected: expected{
				err: nil,
				res: model.SuperWin,
				win: 1000,
			},
		},
		{
			name: "success - win - 1st number is different",
			comb: service.Combination{5, 7, 7},
			bet:  10,
			err:  nil,
			expected: expected{
				err: nil,
				res: model.Win,
				win: 20,
			},
		},
		{
			name: "success - win - 2nd number is different",
			comb: service.Combination{7, 5, 7},
			bet:  10,
			err:  nil,
			expected: expected{
				err: nil,
				res: model.Win,
				win: 20,
			},
		},
		{
			name: "success - win - 3rd number is different",
			comb: service.Combination{7, 7, 5},
			bet:  10,
			err:  nil,
			expected: expected{
				err: nil,
				res: model.Win,
				win: 20,
			},
		},
		{
			name: "success - lose",
			comb: service.Combination{7, 6, 5},
			bet:  10,
			err:  nil,
			expected: expected{
				err: nil,
				res: model.Lose,
				win: -10,
			},
		},
		{
			name: "failure - invalid bet amount",
			comb: service.Combination{0, 0, 0},
			bet:  0,
			err:  service.ErrInvalidBetAmount,
			expected: expected{
				err: service.ErrInvalidBetAmount,
				res: model.SuperWin,
				win: 0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sm := service.NewSlotMachineWithGenerateFn(func() (service.Combination, error) {
				return tc.comb, tc.err
			})

			actualComb, actualWin, actualErr := sm.Spin(tc.bet)

			if actualErr != nil && !errors.Is(actualErr, tc.expected.err) {
				t.Errorf("expected error: %q, got error: %q", tc.expected.err, actualErr)
			}

			if actualComb != tc.comb {
				t.Errorf("expected comb: %q, got comb: %q", tc.comb, actualComb)
			}

			if actualComb.Type() != tc.expected.res {
				t.Errorf(
					"expected outcome type: %q, got outcome type: %q",
					tc.expected.res.String(),
					actualComb.Type(),
				)
			}

			if actualErr == nil && tc.expected.err == nil {
				if actualWin != tc.expected.win {
					t.Errorf("expected win: %d, got win: %d", tc.expected.win, actualWin)
				}
			}
		})
	}
}
