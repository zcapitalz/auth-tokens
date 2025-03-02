package jwtutils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func ParseAndValidateJWTToken(tokenBytes []byte, jwtPrivateKey []byte, signingMethod string) (*jwt.Token, error) {
	token, err := jwt.Parse(string(tokenBytes), func(token *jwt.Token) (interface{}, error) {
		return jwtPrivateKey, nil
	}, jwt.WithValidMethods([]string{signingMethod}))
	if err != nil {
		return nil, errors.Wrap(err, "parse jwt token")
	}
	if !token.Valid {
		return nil, fmt.Errorf("jwt token is not valid")
	}

	return token, nil
}

func GetTimeJWTClaim(claimsMap jwt.MapClaims, claimName string) (time.Time, error) {
	claimAny, err := getJWTClaim(claimsMap, claimName)
	if err != nil {
		return time.Time{}, err
	}
	claim, ok := claimAny.(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("claim is not of type float64")
	}
	sec := int64(claim)

	return time.Unix(sec, 0), nil
}

func GetStringJWTClaim(claimsMap jwt.MapClaims, claimName string) (string, error) {
	claimAny, err := getJWTClaim(claimsMap, claimName)
	if err != nil {
		return "", err
	}
	claim, ok := claimAny.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not of type string", claimName)
	}
	return claim, nil
}

func getJWTClaim(claimsMap jwt.MapClaims, claimName string) (any, error) {
	claim, ok := claimsMap[claimName]
	if !ok {
		return "", fmt.Errorf("claim %s missing", claimName)
	}
	return claim, nil
}
