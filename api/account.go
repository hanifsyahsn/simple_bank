package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/hanifsyahsn/simple_bank/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) createAccount(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusCreated, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(c *gin.Context) {
	var req getAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(c.Request.Context(), req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, account)
}

type getAccountsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=1,max=10"`
	PageID   int32 `form:"page_id,default=1" binding:"min=1"`
}

func (server *Server) getAccounts(c *gin.Context) {
	var req getAccountsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
	}
	limit := req.PageSize
	offset := (req.PageID - 1) * req.PageSize

	arg := db.ListAccountsParams{
		Limit:  limit,
		Offset: offset,
	}

	accounts, err := server.store.ListAccounts(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, accounts)
}
