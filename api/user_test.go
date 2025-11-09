package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/hanifsyahsn/simple_bank/db/mock"
	db "github.com/hanifsyahsn/simple_bank/db/sqlc"
	"github.com/hanifsyahsn/simple_bank/util"
	"github.com/stretchr/testify/require"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckPasswordHash(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)

}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {

	testCases := []struct {
		name             string
		body             createUserRequest
		createUserParams func(req createUserRequest) db.CreateUserParams
		user             func(request createUserRequest) db.User
		buildStubs       func(store *mockdb.MockStore, user db.User, arg db.CreateUserParams, password string)
		checkResponse    func(t *testing.T, recorder *httptest.ResponseRecorder, user db.User)
	}{
		{
			name: "success",
			body: createUserRequest{
				Username: util.RandomOwner(),
				Password: util.RandomString(6),
				FullName: util.RandomOwner(),
				Email:    util.RandomEmail(),
			},
			createUserParams: func(req createUserRequest) db.CreateUserParams {
				return db.CreateUserParams{
					Username: req.Username,
					FullName: req.FullName,
					Email:    req.Email,
				}
			},
			user: func(request createUserRequest) db.User {
				return db.User{
					Username:          request.Username,
					Email:             request.Email,
					FullName:          request.FullName,
					PasswordChangedAt: time.Now(),
					CreatedAt:         time.Now(),
				}
			},
			buildStubs: func(store *mockdb.MockStore, user db.User, arg db.CreateUserParams, password string) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, user db.User) {
				checkResponseMatcher(t, recorder, user)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			requestBody := tc.body

			createUserParams := tc.createUserParams(requestBody)

			user := tc.user(requestBody)

			tc.buildStubs(store, user, createUserParams, requestBody.Password)

			server := newTestServer(t, store)
			require.NotNil(t, server)

			recorder := httptest.NewRecorder()

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			url := fmt.Sprintf("/users")
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder, user)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func checkResponseMatcher(t *testing.T, recorder *httptest.ResponseRecorder, user db.User) {
	data, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, recorder.Code)

	var gotUserResponse userResponse
	err = json.Unmarshal(data, &gotUserResponse)
	fmt.Println(gotUserResponse)
	fmt.Println(user)
	require.Equal(t, user.Username, gotUserResponse.Username)
	require.Equal(t, user.Email, gotUserResponse.Email)
	require.Equal(t, user.FullName, gotUserResponse.FullName)
	require.WithinDuration(t, user.PasswordChangedAt, gotUserResponse.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user.CreatedAt, gotUserResponse.CreatedAt, time.Second)
}
