package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
// embed queries into Store, this is like inheritance in Golang, it's called composition
type Store struct {
	*Queries

	db *sql.DB
}

// NewStore constructor
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer      Transfer `json:"transfer"`
	FromAccountID int64    `json:"from_account_id"`
	ToAccountID   int64    `json:"to_account_id"`
	FromEntry     Entry    `json:"from_entry"`
	ToEntry       Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another
// It creates a transfer record, add new account entries, updates account balance with single database transaction
func (s *Store) TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {

		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: args.FromAccountID,
			ToAccountID:   args.ToAccountID,
			Amount:        args.Amount,
		})

		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.FromAccountID,
			Amount:    -args.Amount,
		})

		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.ToAccountID,
			Amount:    +args.Amount,
		})

		return nil
	})

	// TODO: Update both accounts balance

	return result, err
}

// executes a function within a database transaction
func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)

	if err != nil {
		// rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}

		return err
	}

	// all operations are successful
	return tx.Commit()
}
