package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStore_TransferTx(t *testing.T) {
	store := NewStore(testDB)

	sender := createRandomAccount(t)
	receiver := createRandomAccount(t)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	// create a channel of errors to publish to errors to the main thread from the go routine
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {

		// run on a separate thread, different go routine
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: sender.ID,
				ToAccountID:   receiver.ID,
				Amount:        amount,
			})

			// publish the values to the channel
			errs <- err
			results <- result
		}()
	}

	exists := make(map[int]bool)

	// check results
	for i := 0; i < n; i++ {

		// receive the errors
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, sender.ID, transfer.FromAccountID)
		require.Equal(t, receiver.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, sender.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, receiver.ID, toEntry.AccountID)
		require.Equal(t, +amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, sender.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, receiver.ID, toAccount.ID)

		// check accounts balance
		diff1 := sender.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - receiver.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)

		require.NotContains(t, exists, k)
		exists[k] = true
	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), sender.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), receiver.ID)
	require.NoError(t, err)

	require.Equal(t, sender.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, receiver.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestStore_TransferTx_Deadlock(t *testing.T) {
	store := NewStore(testDB)

	sender := createRandomAccount(t)
	receiver := createRandomAccount(t)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	// create a channel of errors to publish to errors to the main thread from the go routine
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := sender.ID
		toAccountID := receiver.ID

		// change the sender and receiver if the loop variable is odd
		// we are trying to create a deadlock scenario here
		if i%2 == 1 {
			fromAccountID = receiver.ID
			toAccountID = sender.ID
		}

		// run on a separate thread, different go routine
		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			// publish the values to the channel
			errs <- err
		}()
	}

	// check results
	for i := 0; i < n; i++ {

		// receive the errors
		err := <-errs
		require.NoError(t, err)

	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), sender.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), receiver.ID)
	require.NoError(t, err)

	// because they are receiving and sending the same amount of money to each other
	require.Equal(t, sender.Balance, updatedAccount1.Balance)
	require.Equal(t, receiver.Balance, updatedAccount2.Balance)
}
