package api

import (
	db "github.com/aybarsacar/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves all http requests for banking service
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer constructor
func NewServer(store db.Store) *Server {
	server := Server{
		store: store,
	}

	router := gin.Default()

	// register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	// add routes
	router.POST("/api/v1/accounts", server.createAccount)
	router.GET("/api/v1/accounts/:id", server.getAccount)
	router.GET("/api/v1/accounts", server.listAccounts)

	router.POST("/api/v1/transfers", server.createTransfer)

	server.router = router

	return &server
}

// Start runs the HTTP server ona specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}