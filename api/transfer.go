package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "github.com/aybarsacar/simplebank/db/sqlc"
	"github.com/aybarsacar/simplebank/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		// user sent invalid data, send response
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// insert new account into the database
	args := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	fromAccount, isValid := server.validAccount(ctx, req.FromAccountID, req.Currency)

	if !isValid {
		return
	}

	// get the token payload from the middleware
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account does not belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if _, isValid := server.validAccount(ctx, req.ToAccountID, req.Currency); !isValid {
		return
	}

	result, err := server.store.TransferTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// account is successfully created - send account back to client
	ctx.JSON(http.StatusOK, result)
}

// account with a specific id exists and currency matches the input currency
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {

	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {

		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
