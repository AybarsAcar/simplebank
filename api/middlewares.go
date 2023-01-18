package api

import (
	"errors"
	"github.com/aybarsacar/simplebank/token"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// higher order function that will return the authentication middleware function
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	// this is the actual authentication middleware
	return func(context *gin.Context) {
		authorizationHeader := context.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) <= 0 {
			err := errors.New("authorization header is not provided")
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorisation header format")
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := errors.New("unsupported authorisation type")
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// store the payload in the context with key
		// we can access this payload in the next handlers in the request pipeline
		context.Set(authorizationPayloadKey, payload)

		// go to the next handler
		context.Next()
	}
}
