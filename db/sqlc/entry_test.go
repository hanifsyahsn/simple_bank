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

func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    account.Balance + util.RandomMoney(),
	}
	createEntry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, createEntry)

	require.Equal(t, createEntry.AccountID, account.ID)
	require.Equal(t, createEntry.Amount, arg.Amount)

	require.NotZero(t, createEntry.ID)
	require.NotZero(t, createEntry.CreatedAt)

	return createEntry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry := createRandomEntry(t, account)

	getEntry, err := testQueries.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, getEntry)

	require.Equal(t, entry.ID, getEntry.ID)
	require.Equal(t, entry.AccountID, getEntry.AccountID)
	require.Equal(t, entry.Amount, getEntry.Amount)
	require.WithinDuration(t, entry.CreatedAt, getEntry.CreatedAt, time.Second)
}

func TestGetListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Offset:    5,
		Limit:     5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	require.Len(t, entries, 5)
	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

func TestGetListEntries_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	queries := New(db)

	account := createRandomAccount(t)

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	mock.ExpectQuery("SELECT id, account_id, amount, created_at FROM entries").
		WillReturnError(fmt.Errorf("query error"))

	_, err = queries.ListEntries(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query error")

	rows := mock.NewRows([]string{"id", "account_id", "amount", "created_at"}).
		AddRow("1", "theId", "theAmount", "theCreatedAt")
	mock.ExpectQuery("SELECT id, account_id, amount, created_at FROM entries").
		WillReturnRows(rows)

	_, err = queries.ListEntries(context.Background(), arg)
	require.Error(t, err)

	rows = mock.NewRows([]string{"id", "account_id", "amount", "created_at"}).
		AddRow(1, account.ID, 100, time.Now()).
		RowError(0, fmt.Errorf("iteration error"))
	mock.ExpectQuery("SELECT id, account_id, amount, created_at FROM entries").
		WillReturnRows(rows)

	_, err = queries.ListEntries(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "iteration error")
}
