package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/hanifsyahsn/simple_bank/db/mock"
	db "github.com/hanifsyahsn/simple_bank/db/sqlc"
	"github.com/hanifsyahsn/simple_bank/token"
	"github.com/hanifsyahsn/simple_bank/util"
	"github.com/stretchr/testify/require"
)

func TestTransferAPI(t *testing.T) {
	fromAccount := db.Account{
		ID:       1,
		Owner:    "A",
		Currency: util.USD,
		Balance:  100,
	}
	toAccount := db.Account{
		ID:       2,
		Owner:    "B",
		Currency: util.USD,
		Balance:  100,
	}
	fromAccountCAD := db.Account{
		ID:       3,
		Owner:    "C",
		Currency: util.CAD,
		Balance:  100,
	}
	toAccountCAD := db.Account{
		ID:       4,
		Owner:    "D",
		Currency: util.CAD,
		Balance:  100,
	}
	transferAmount := util.RandomInt(1, 50)

	fromAccountEntry := createEntry(fromAccount.ID, transferAmount)
	toAccountEntry := createEntry(toAccount.ID, -transferAmount)

	transfer := createTransfer(fromAccount.ID, toAccount.ID, transferAmount)

	testCases := []struct {
		name          string
		arg           db.TransferTxParams
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		reqBody       transferRequest
		expectResp    db.TransferTxResult
		buildStub     func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult)
	}{
		{
			name: "success",
			arg: db.TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{
				Transfer:    transfer,
				ToAccount:   toAccount,
				FromAccount: fromAccount,
				ToEntry:     toAccountEntry,
				FromEntry:   fromAccountEntry,
			},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(expRes, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, expRes)
			},
		},
		{
			name: "UnsupportedCurrency",
			arg: db.TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      "IDR",
			},
			expectResp: db.TransferTxResult{
				Transfer:    transfer,
				ToAccount:   toAccount,
				FromAccount: fromAccount,
				ToEntry:     toAccountEntry,
				FromEntry:   fromAccountEntry,
			},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrencyFromAccount",
			arg:  db.TransferTxParams{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccountCAD.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccountCAD.ID)).Times(1).Return(fromAccountCAD, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				var resp map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Contains(t, resp["error"], "invalid currency")
			},
		},
		{
			name: "InvalidCurrencyToAccount",
			arg:  db.TransferTxParams{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccountCAD.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountCAD.ID)).Times(1).Return(toAccountCAD, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

				var resp map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Contains(t, resp["error"], "invalid currency")
			},
		},
		{
			name: "InternalErrorOnTransferTx",
			arg: db.TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "AccountNotFound",
			arg:  db.TransferTxParams{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnGetAccount",
			arg:  db.TransferTxParams{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			reqBody: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        transferAmount,
				Currency:      util.USD,
			},
			expectResp: db.TransferTxResult{},
			buildStub: func(store *mockdb.MockStore, arg db.TransferTxParams, expRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, expRes db.TransferTxResult) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			server := newTestServer(t, store)
			require.NotNil(t, server)

			body, err := json.Marshal(&testCase.reqBody)
			require.NoError(t, err)

			testCase.buildStub(store, testCase.arg, testCase.expectResp)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/transfers")
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)

			testCase.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder, testCase.expectResp)
		})
	}

}

func createEntry(accountID, amount int64) (res db.Entry) {
	res = db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: accountID,
		Amount:    amount,
	}
	return
}

func createTransfer(fromAccountID, toAccountID int64, amount int64) (res db.Transfer) {
	res = db.Transfer{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
	}
	return
}

//func toCreateAccount(id, currency, balance int64, owner string) (res db.Account) {
//	res = db.Account{
//		ID: id | 1,
//	}
//	return
//}

func requireBodyMatchTransfer(t *testing.T, buffer *bytes.Buffer, expected db.TransferTxResult) {
	data, err := io.ReadAll(buffer)
	require.NoError(t, err)

	var transferResult db.TransferTxResult
	err = json.Unmarshal(data, &transferResult)
	require.NoError(t, err)

	require.Equal(t, expected.Transfer.FromAccountID, transferResult.Transfer.FromAccountID)
	require.Equal(t, expected.Transfer.ToAccountID, transferResult.Transfer.ToAccountID)
	require.Equal(t, expected.Transfer.Amount, transferResult.Transfer.Amount)
	require.Equal(t, expected.FromAccount, transferResult.FromAccount)
	require.Equal(t, expected.ToAccount, transferResult.ToAccount)
	require.Equal(t, expected.FromEntry, transferResult.FromEntry)
	require.Equal(t, expected.ToEntry, transferResult.ToEntry)
}
