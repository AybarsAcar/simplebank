package db

import (
	"context"
	"database/sql"
	"github.com/aybarsacar/simplebank/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueries_CreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestQueries_GetAccount(t *testing.T) {
	// create account
	account1 := createRandomAccount(t)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)

	// they are both created within 1 second, we decide on the accuracy
	// they should be exactly the same as we are getting the account
	// get action is a command not a query, so it should not change data
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, 1)
}

func TestQueries_UpdateAccount(t *testing.T) {
	account1 := createRandomAccount(t)

	args := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}

	account2, err := testQueries.UpdateAccount(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Currency, account2.Currency)

	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, 1)
}

func TestQueries_DeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	require.Empty(t, account2)
}

func TestQueries_ListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	// skip first 5 and return the next 5
	args := ListAccountsParams{Limit: 5, Offset: 5}

	accounts, err := testQueries.ListAccounts(context.Background(), args)

	require.NoError(t, err)

	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
	}
}

func createRandomAccount(t *testing.T) Account {
	args := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), args)

	require.NoError(t, err)

	require.NotEmpty(t, account)

	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}
