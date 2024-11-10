package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/storage"
	"github.com/apolunin/slotgame/logger"
	"log/slog"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid login/password")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

type (
	userServiceRepo interface {
		CreateUser(
			ctx context.Context,
			login string,
			password string,
			firstName string,
			lastName string,
			balance int64,
		) (*model.User, error)

		GetUserByLogin(
			ctx context.Context,
			login string,
		) (*model.User, error)

		GetUserBalanceByLogin(
			ctx context.Context,
			login string,
		) (int64, error)

		SetUserBalanceByLogin(
			ctx context.Context,
			login string,
			newBalance int64,
		) error

		RunTx(ctx context.Context, f storage.TxFn) (err error)
	}

	userServiceAuth interface {
		HashPassword(password string) (string, error)
		IsValidPassword(hash, password string) bool
		CreateToken(login string) (string, error)
	}

	UserService struct {
		storage userServiceRepo
		auth    userServiceAuth
	}
)

func NewUserService(storage userServiceRepo, auth userServiceAuth) *UserService {
	return &UserService{
		storage: storage,
		auth:    auth,
	}
}

func (us *UserService) CreateUser(
	ctx context.Context,
	login string,
	password string,
	firstName string,
	lastName string,
	balance int64,
) (*model.User, error) {
	log := slog.With(logger.FieldUser, login)

	log.Debug("user creation started")
	defer log.Debug("user creation completed")

	hashedPwd, err := us.auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password for user %s", login)
	}

	return us.storage.CreateUser(ctx, login, hashedPwd, firstName, lastName, balance)
}

func (us *UserService) Login(
	ctx context.Context,
	login string,
	password string,
) (string, error) {
	log := slog.With(logger.FieldUser, login)

	log.Debug("user login started")
	defer log.Debug("user login completed")

	usr, err := us.storage.GetUserByLogin(ctx, login)

	switch {
	case err != nil:
		return "", fmt.Errorf("failed to fetch user by login %q", login)
	case err == nil && usr == nil:
		return "", fmt.Errorf("%w: %q", ErrUserNotFound, login)
	case !us.auth.IsValidPassword(usr.Password, password):
		return "", fmt.Errorf("%w, login: %q", ErrInvalidCredentials, login)
	}

	token, err := us.auth.CreateToken(login)
	if err != nil {
		return "", fmt.Errorf("failed to create jwt token for user %q", login)
	}

	return token, err
}

func (us *UserService) GetUserByLogin(
	ctx context.Context,
	login string,
) (*model.User, error) {
	log := slog.With(logger.FieldUser, login)

	log.Debug("getting user by login started")
	defer log.Debug("getting user by login completed")

	switch usr, err := us.storage.GetUserByLogin(ctx, login); {
	case err != nil:
		return nil, fmt.Errorf("failed to fetch user by login %q", login)
	case err == nil && usr == nil:
		return nil, fmt.Errorf("%w: %q", ErrUserNotFound, login)
	default:
		return usr, nil
	}
}

func (us *UserService) DepositFunds(
	ctx context.Context,
	login string,
	amount int64,
) (int64, error) {
	log := slog.With(logger.FieldUser, login)

	log.Debug("funds deposit started")
	defer log.Debug("funds deposit completed")

	var balance int64

	if err := us.storage.RunTx(ctx, func(ctx context.Context) error {
		currentBalance, err := us.storage.GetUserBalanceByLogin(ctx, login)

		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return ErrUserNotFound
		case err != nil:
			return err
		}

		balance = currentBalance + amount

		switch err = us.storage.SetUserBalanceByLogin(ctx, login, balance); {
		case errors.Is(err, storage.ErrUserNotFound):
			return ErrUserNotFound
		case err != nil:
			return err
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return balance, nil
}

func (us *UserService) WithdrawFunds(
	ctx context.Context,
	login string,
	amount int64,
) (int64, error) {
	log := slog.With(logger.FieldUser, login)

	log.Debug("funds withdrawal started")
	defer log.Debug("funds withdrawal completed")

	var balance int64

	if err := us.storage.RunTx(ctx, func(ctx context.Context) error {
		currentBalance, err := us.storage.GetUserBalanceByLogin(ctx, login)

		switch {
		case errors.Is(err, storage.ErrUserNotFound):
			return ErrUserNotFound
		case err != nil:
			return err
		}

		if currentBalance < amount {
			return ErrInsufficientFunds
		}

		balance = currentBalance - amount

		switch err = us.storage.SetUserBalanceByLogin(ctx, login, balance); {
		case errors.Is(err, storage.ErrUserNotFound):
			return ErrUserNotFound
		case err != nil:
			return err
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return balance, nil
}
