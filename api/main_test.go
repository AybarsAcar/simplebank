package api

import (
	"github.com/gin-gonic/gin"
	"os"
	"testing"
)

// configure Gin to run in test mode
// default is running in debug mode

// main entry point to the tests
func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	// start running the unit tests
	os.Exit(m.Run())
}
