package db

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hanifsyahsn/simple_bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomTransfer(t *testing.T, account1, account2 Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, transfer.FromAccountID, account1.ID)
	require.Equal(t, transfer.ToAccountID, account2.ID)
	require.Equal(t, transfer.Amount, arg.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	createRandomTransfer(t, account1, account2)
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer := createRandomTransfer(t, account1, account2)

	transferResult, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transferResult)

	require.Equal(t, transferResult.ID, transfer.ID)
	require.Equal(t, transferResult.FromAccountID, transfer.FromAccountID)
	require.Equal(t, transferResult.ToAccountID, transfer.ToAccountID)
	require.Equal(t, transferResult.Amount, transfer.Amount)

	require.WithinDuration(t, transferResult.CreatedAt, transfer.CreatedAt, time.Second)
}

func TestGetListTransfers(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, account1, account2)
	}

	arg := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	}

	transfersResult, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfersResult)
	require.Len(t, transfersResult, 5)

	for _, transfer := range transfersResult {
		require.NotEmpty(t, transfer)
	}
}

func TestGetListTransfers_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	queries := New(db)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	arg := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	}

	mock.ExpectQuery("SELECT id, from_account_id, to_account_id, amount, created_at FROM transfers").
		WillReturnError(fmt.Errorf("query error"))

	_, err = queries.ListTransfers(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query error")

	rows := mock.NewRows([]string{"id", "from_account_id", "to_account_id", "amount", "created_at"}).
		AddRow("not-an-int", account1.ID, account2.ID, "not-a-number", "not-a-time")

	mock.ExpectQuery("SELECT id, from_account_id, to_account_id, amount, created_at FROM transfers").
		WillReturnRows(rows)

	_, err = queries.ListTransfers(context.Background(), arg)
	require.Error(t, err)

	rows = mock.NewRows([]string{"id", "from_account_id", "to_account_id", "amount", "created_at"}).
		AddRow(1, account1.ID, account2.ID, 100, time.Now()).
		RowError(0, fmt.Errorf("iteration error"))

	mock.ExpectQuery("SELECT id, from_account_id, to_account_id, amount, created_at FROM transfers").
		WillReturnRows(rows)

	_, err = queries.ListTransfers(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "iteration error")
}
