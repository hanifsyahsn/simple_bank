package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hanifsyahsn/simple_bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)
	accountResult, err := testQueries.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountResult)

	require.Equal(t, account.ID, accountResult.ID)
	require.Equal(t, account.Owner, accountResult.Owner)
	require.Equal(t, account.Balance, accountResult.Balance)
	require.Equal(t, account.Currency, accountResult.Currency)
	require.WithinDuration(t, account.CreatedAt, accountResult.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)
	arg := UpdateAccountParams{
		ID:      account.ID,
		Balance: util.RandomMoney(),
	}
	accountUpdate, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accountUpdate)

	require.Equal(t, account.ID, accountUpdate.ID)
	require.Equal(t, accountUpdate.Owner, account.Owner)
	require.Equal(t, accountUpdate.Balance, arg.Balance)
	require.Equal(t, accountUpdate.Currency, account.Currency)
	require.WithinDuration(t, account.CreatedAt, accountUpdate.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)
	accountDeletedError := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, accountDeletedError)

	getAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, getAccount)
}

func TestGetListAccounts(t *testing.T) {
	var lastAccount Account
	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}

func TestGetListAccounts_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	queries := New(db)

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	mock.ExpectQuery("SELECT id, owner, balance, currency, created_at FROM accounts").
		WillReturnError(fmt.Errorf("query error"))

	_, err = queries.ListAccounts(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query error")

	rows := sqlmock.NewRows([]string{"id", "owner", "balance", "currency", "created_at"}).
		AddRow("1", "ownerName", "theBalance", "theCurrency", "theCreatedAt")
	mock.ExpectQuery("SELECT id, owner, balance, currency, created_at FROM accounts").
		WillReturnRows(rows)

	_, err = queries.ListAccounts(context.Background(), arg)
	require.Error(t, err)

	rows = sqlmock.NewRows([]string{"id", "owner", "balance", "currency", "created_at"}).
		AddRow(1, "owner", 100, "USD", time.Now()).
		RowError(0, fmt.Errorf("iteration error"))

	mock.ExpectQuery("SELECT id, owner, balance, currency, created_at FROM accounts").
		WillReturnRows(rows)

	_, err = queries.ListAccounts(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "iteration error")
}
