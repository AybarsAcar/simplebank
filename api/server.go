package api

import (
	"fmt"
	db "github.com/aybarsacar/simplebank/db/sqlc"
	"github.com/aybarsacar/simplebank/token"
	"github.com/aybarsacar/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves all http requests for banking service
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer constructor
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRoutes()

	return &server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()

	router.POST("/api/v1/users", server.createUser)
	router.POST("/api/v1/users/login", server.loginUser)

	// create auth middleware, every request that needs to get JWT Payload and
	// authenticate is added to this route now
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/api/v1/accounts", server.createAccount)
	authRoutes.GET("/api/v1/accounts/:id", server.getAccount)
	authRoutes.GET("/api/v1/accounts", server.listAccounts)

	authRoutes.POST("/api/v1/transfers", server.createTransfer)

	server.router = router
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
