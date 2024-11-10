package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type (
	Storage struct {
		conn *pgxpool.Pool
	}

	txKey struct{}
	TxFn  func(ctx context.Context) error
)

func NewStorage(conn *pgxpool.Pool) *Storage {
	return &Storage{
		conn: conn,
	}
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

func (s *Storage) CreateUser(
	ctx context.Context,
	login string,
	password string,
	firstName string,
	lastName string,
	balance int64,
) (*model.User, error) {
	usr := &model.User{
		FirstName: firstName,
		LastName:  lastName,
		Login:     login,
		Password:  password,
		Balance:   balance,
	}

	if err := s.RunTx(ctx, func(ctx context.Context) error {
		tx := txFromContext(ctx)

		switch u, err := s.GetUserByLogin(ctx, login); {
		case err != nil:
			return fmt.Errorf("failed to check if user already exists: %w", err)
		case u != nil:
			return fmt.Errorf("cannot create user %q: %w", usr.Login, ErrUserExists)
		}

		query := `
			INSERT INTO "user" (first_name, last_name, login, password, balance) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING id
		`

		return tx.QueryRow(ctx, query, firstName, lastName, login, password, balance).Scan(&usr.ID)
	}); err != nil {
		return nil, err
	}

	return usr, nil
}

func (s *Storage) GetUserByLogin(
	ctx context.Context,
	login string,
) (*model.User, error) {
	usr := &model.User{}

	if err := s.RunTx(ctx, func(ctx context.Context) error {
		tx := txFromContext(ctx)

		query := `
			SELECT id, first_name, last_name, login, password, balance
			FROM "user" WHERE login = $1
		`

		err := tx.QueryRow(ctx, query, login).Scan(
			&usr.ID,
			&usr.FirstName,
			&usr.LastName,
			&usr.Login,
			&usr.Password,
			&usr.Balance,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			usr = nil
			err = nil
		}

		return err
	}); err != nil {
		return nil, err
	}

	return usr, nil
}

func (s *Storage) GetUserBalanceByLogin(
	ctx context.Context,
	login string,
) (int64, error) {
	var balance int64

	if err := s.RunTx(ctx, func(ctx context.Context) error {

		var (
			query = `SELECT balance FROM "user" WHERE login = $1`
			tx    = txFromContext(ctx)
			err   = tx.QueryRow(ctx, query, login).Scan(&balance)
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}); err != nil {
		return 0, err
	}

	return balance, nil
}

func (s *Storage) SetUserBalanceByLogin(
	ctx context.Context,
	login string,
	newBalance int64,
) error {
	return s.RunTx(ctx, func(ctx context.Context) error {
		var (
			query    = `UPDATE "user" SET balance = $1 WHERE login = $2`
			tx       = txFromContext(ctx)
			cmd, err = tx.Exec(ctx, query, newBalance, login)
		)

		switch {
		case err != nil:
			return fmt.Errorf("error updating user balance: %w", err)
		case cmd.RowsAffected() == 0:
			return ErrUserNotFound
		}

		return nil
	})
}

func (s *Storage) CreateSpin(
	ctx context.Context,
	userID string,
	combination string,
	result model.SpinResult,
	betAmount int64,
	winAmount int64,
) (*model.Spin, error) {
	spin := &model.Spin{
		UserID:      userID,
		Combination: combination,
		Result:      result,
		BetAmount:   betAmount,
		WinAmount:   winAmount,
	}

	if err := s.RunTx(ctx, func(ctx context.Context) error {
		var (
			tx    = txFromContext(ctx)
			query = `
				INSERT INTO "spin_result" (user_id, combination, result, bet_amount, win_amount)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, created_at
			`
		)

		return tx.QueryRow(
			ctx,
			query,
			userID,
			combination,
			result,
			betAmount,
			winAmount,
		).Scan(&spin.ID, &spin.CreatedAt)
	}); err != nil {
		return nil, err
	}

	return spin, nil
}

func (s *Storage) GetSpinResults(
	ctx context.Context,
	userID string,
	limit int64,
	offset int64,
) ([]*model.Spin, error) {
	var results []*model.Spin

	if err := s.RunTx(ctx, func(ctx context.Context) error {
		var (
			tx    = txFromContext(ctx)
			query = `
				SELECT id, user_id, combination, result, bet_amount, win_amount, created_at
				FROM spin_result
				WHERE user_id = $1
				ORDER BY created_at DESC
				LIMIT $2 OFFSET $3;
			`

			rows, err = tx.Query(ctx, query, userID, limit, offset)
		)

		if err != nil {
			return fmt.Errorf("failed to get spin results: %w", err)
		}

		defer rows.Close()

		for rows.Next() {
			var spin model.Spin

			if err := rows.Scan(
				&spin.ID,
				&spin.UserID,
				&spin.Combination,
				&spin.Result,
				&spin.BetAmount,
				&spin.WinAmount,
				&spin.CreatedAt,
			); err != nil {
				return fmt.Errorf("failed scanning row: %w", err)
			}

			results = append(results, &spin)
		}

		if rows.Err() != nil {
			return fmt.Errorf("rows iteration error: %w", rows.Err())
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Storage) RunTx(ctx context.Context, f TxFn) (err error) {
	var (
		tx      = txFromContext(ctx)
		txOwner = false
	)

	if tx == nil {
		tx, err = s.conn.Begin(ctx)

		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		txOwner = true
		ctx = context.WithValue(ctx, txKey{}, tx)
	}

	if err = f(ctx); err != nil {
		if txOwner {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				slog.With(logger.FieldError, rbErr).Error("failed to rollback transaction")
			}
		}

		return err
	}

	if txOwner {
		return tx.Commit(ctx)
	}

	return nil
}

func txFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return nil
}
