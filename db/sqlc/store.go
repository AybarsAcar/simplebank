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
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account_id"`
	ToAccount   Account  `json:"to_account_id"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
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

		// always update the account with the smaller id first to avoid deadlocks
		if args.FromAccountID < args.ToAccountID {
			// decrement the from accounts balance by the amount
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, args.FromAccountID, -args.Amount, args.ToAccountID, args.Amount)
		} else {
			// decrement the from accounts balance by the amount
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, args.ToAccountID, args.Amount, args.FromAccountID, -args.Amount)
		}

		return nil
	})

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

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {

	account1, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})

	if err != nil {
		return account1, account2, err
	}

	account2, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})

	return
}
