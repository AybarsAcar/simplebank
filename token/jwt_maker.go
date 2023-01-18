package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const minSecretKeySize = 32

// JWTMaker is a JSON web token maker
type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

// implement the interface

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {

	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken checks if the token is valid or not
// if valid returns the payload stored inside the body of token
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {

	// parse the token
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			// algorithm or token does not match with our signing algorithm
			return nil, ErrInvalidToken
		}

		// algorithm matches

		return []byte(maker.secretKey), nil
	})

	if err != nil {
		// token is invalid or it is expired
		verr, ok := err.(*jwt.ValidationError)

		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	// successfully parsed -> get the payload
	payload, ok := jwtToken.Claims.(*Payload)

	if !ok {
		return nil, ErrInvalidToken
	}

	// payload got successfully
	return payload, nil
}
