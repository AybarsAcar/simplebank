package api

import (
	db "github.com/aybarsacar/simplebank/db/sqlc"
	"github.com/aybarsacar/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

// configure Gin to run in test mode
// default is running in debug mode

func newTestServer(t *testing.T, store db.Store) *Server {

	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

// main entry point to the tests
func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	// start running the unit tests
	os.Exit(m.Run())
}
