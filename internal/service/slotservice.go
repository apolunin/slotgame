package service

import (
	"context"
	"errors"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/storage"
)

type (
	slotServiceRepo interface {
		RunTx(ctx context.Context, f storage.TxFn) (err error)

		GetUserBalanceByLogin(
			ctx context.Context,
			login string,
		) (int64, error)

		SetUserBalanceByLogin(
			ctx context.Context,
			login string,
			newBalance int64,
		) error

		CreateSpin(
			ctx context.Context,
			userID string,
			combination string,
			result model.SpinResult,
			betAmount int64,
			winAmount int64,
		) (*model.Spin, error)

		GetSpinResults(
			ctx context.Context,
			userID string,
			limit int64,
			offset int64,
		) ([]*model.Spin, error)
	}

	SlotService struct {
		storage     slotServiceRepo
		slotMachine *SlotMachine
	}
)

func NewSlotService(
	storage slotServiceRepo,
	slotMachine *SlotMachine,
) *SlotService {
	return &SlotService{
		storage:     storage,
		slotMachine: slotMachine,
	}
}

func (ss *SlotService) Spin(
	ctx context.Context,
	user *model.User,
	betAmount int64,
) (*model.Spin, int64, error) {
	var (
		newBalance int64
		spin       *model.Spin
	)

	if err := ss.storage.RunTx(ctx, func(ctx context.Context) error {
		currentBalance, err := ss.storage.GetUserBalanceByLogin(ctx, user.Login)

		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return ErrUserNotFound
		case err != nil:
			return err
		}

		if currentBalance < betAmount {
			return ErrInsufficientFunds
		}

		combination, winAmount, err := ss.slotMachine.Spin(betAmount)
		if err != nil {
			return err
		}

		newBalance = currentBalance + winAmount

		if err := ss.storage.SetUserBalanceByLogin(
			ctx,
			user.Login,
			newBalance,
		); err != nil {
			return err
		}

		spin, err = ss.storage.CreateSpin(
			ctx,
			user.ID,
			combination.String(),
			combination.Type(),
			betAmount,
			winAmount,
		)

		return err
	}); err != nil {
		return nil, 0, err
	}

	return spin, newBalance, nil
}

func (ss *SlotService) GetSpinResults(
	ctx context.Context,
	user *model.User,
	limit int64,
	offset int64,
) ([]*model.Spin, error) {
	return ss.storage.GetSpinResults(ctx, user.ID, limit, offset)
}
