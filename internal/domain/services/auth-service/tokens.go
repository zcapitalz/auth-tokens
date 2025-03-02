package authservice

import (
	"auth/internal/domain"
	jwtutils "auth/internal/utils/jwt-utils"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	accessTokenJWTSigningMethod = jwt.SigningMethodHS512
)

const (
	UserIDJWTClaimName         = "sub"
	UserIPJWTClaimName         = "sub_ip"
	RefreshTokenIDJWTClaimName = "refresh_token_id"
	ExpirationTimeJWTClaimName = "exp"
)

func parseAccessTokenFromJWT(token *jwt.Token) (*domain.AccessToken, error) {
	var accessToken domain.AccessToken
	if claimsMap, ok := token.Claims.(jwt.MapClaims); ok {
		userIDStr, err := jwtutils.GetStringJWTClaim(claimsMap, UserIDJWTClaimName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", UserIDJWTClaimName))
		}
		accessToken.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", UserIDJWTClaimName))
		}
		accessToken.UserIP, err = jwtutils.GetStringJWTClaim(claimsMap, UserIPJWTClaimName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", UserIPJWTClaimName))
		}
		refreshTokenID, err := jwtutils.GetStringJWTClaim(claimsMap, RefreshTokenIDJWTClaimName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", RefreshTokenIDJWTClaimName))
		}
		accessToken.RefreshTokenID, err = uuid.Parse(refreshTokenID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", RefreshTokenIDJWTClaimName))
		}
		accessToken.ExpTime, err = jwtutils.GetTimeJWTClaim(claimsMap, ExpirationTimeJWTClaimName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("parse %s claim", ExpirationTimeJWTClaimName))
		}

		return &accessToken, nil
	}

	return nil, fmt.Errorf("claim missing")
}

func generateRefreshTokenValueBytes() ([]byte, error) {
	refreshUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "create uuid")
	}
	refreshBytes, err := refreshUUID.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal uuid")
	}

	return refreshBytes, nil
}
