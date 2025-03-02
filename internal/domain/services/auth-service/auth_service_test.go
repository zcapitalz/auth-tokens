package authservice

import (
	"testing"
	"time"

	"auth/internal/domain"
	"auth/internal/domain/services/auth-service/mocks"
	jwtutils "auth/internal/utils/jwt-utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtPrivateKey        = []byte("private-key")
	accessTokenDuration  = time.Hour * 2
	refreshTokenDuration = time.Hour * 12

	userID = uuid.MustParse("8798e65e-dc84-4a7d-879e-2a52e67d86da")
	userIP = "127.0.0.1"

	refreshTokenID        = uuid.MustParse("3e02eeb9-de9a-4e0a-857b-1293c25bd776")
	refreshTokenValue     = []byte{71, 34, 18, 186, 54, 175, 79, 64, 150, 16, 134, 201, 147, 39, 67, 45}
	refreshTokenValueHash = MustGenerateBcryptHashFromPassword(refreshTokenValue, bcrypt.DefaultCost)
	refreshToken          = domain.RefreshToken{
		ID:             refreshTokenID,
		ValueHash:      refreshTokenValueHash,
		ExpirationTime: time.Now().Add(refreshTokenDuration)}

	accessTokenJWT = jwt.NewWithClaims(accessTokenJWTSigningMethod,
		jwt.MapClaims{
			UserIDJWTClaimName:         userID.String(),
			UserIPJWTClaimName:         userIP,
			ExpirationTimeJWTClaimName: time.Now().Add(accessTokenDuration).Unix(),
			RefreshTokenIDJWTClaimName: refreshTokenID,
		})
	accessTokenSigned = []byte(mustSignJWTString(accessTokenJWT, jwtPrivateKey))

	session = &domain.Session{
		AccessTokenSigned: accessTokenSigned,
		RefreshTokenValue: refreshTokenValue,
	}
)

func TestCreateSession_Success(t *testing.T) {
	service, refreshTokenRepository, _ := newServiceAndMocks(t)
	refreshTokenRepository.
		On("Create", mock.AnythingOfType("*domain.RefreshToken")).
		Return(refreshTokenID, nil)
	startTime := time.Now().Truncate(time.Second) // truncate time since jwt claim "exp" truncates it to seconds

	session, err := service.CreateSession(userID, userIP)

	assert.NoError(t, err)
	accessTokenJWT, err := jwtutils.ParseAndValidateJWTToken(
		session.AccessTokenSigned, jwtPrivateKey,
		accessTokenJWTSigningMethod.Name)
	assert.NoError(t, err)
	claimsMap, ok := accessTokenJWT.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	userIDClaim, err := jwtutils.GetStringJWTClaim(claimsMap, UserIDJWTClaimName)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), userIDClaim)
	userIPClaim, err := jwtutils.GetStringJWTClaim(claimsMap, UserIPJWTClaimName)
	assert.NoError(t, err)
	assert.Equal(t, userIP, userIPClaim)
	accessExpTime, err := jwtutils.GetTimeJWTClaim(claimsMap, ExpirationTimeJWTClaimName)
	assert.NoError(t, err)
	assert.True(t, !accessExpTime.Before(startTime.Add(accessTokenDuration))) // use !Before instead of After because time is truncated to seconds and two values can be equal
	refreshTokenIDClaim, err := jwtutils.GetStringJWTClaim(claimsMap, RefreshTokenIDJWTClaimName)
	assert.NoError(t, err)
	_, err = uuid.Parse(refreshTokenIDClaim)
	assert.NoError(t, err)

	var parsedRefreshTokenValue uuid.UUID
	err = (&parsedRefreshTokenValue).UnmarshalBinary(session.RefreshTokenValue)
	assert.NoError(t, err)
}

func TestRefreshSession_Success(t *testing.T) {
	service, refreshTokenRepository, _ := newServiceAndMocks(t)

	refreshTokenRepository.
		On("GetByID", refreshToken.ID).
		Return(&refreshToken, nil)
	refreshTokenRepository.
		On("DeleteByID", refreshToken.ID).
		Return(nil)
	refreshTokenRepository.
		On("Create", mock.AnythingOfType("*domain.RefreshToken")).
		Return(uuid.New(), nil)

	_, err := service.RefreshSession(session, userIP)
	assert.NoError(t, err)
}

func TestRefreshSession_WrongRefreshToken(t *testing.T) {
	service, refreshTokenRepository, _ := newServiceAndMocks(t)

	refreshTokenRepository.
		On("GetByID", refreshToken.ID).
		Return(&refreshToken, nil)

	var unauthorizedError *domain.UnauthorizedError
	_, err := service.RefreshSession(
		&domain.Session{
			AccessTokenSigned: session.AccessTokenSigned,
			RefreshTokenValue: append(session.RefreshTokenValue, 'a'),
		},
		userIP)
	assert.ErrorAs(t, err, &unauthorizedError)
}

func TestRefreshSession_RefreshTokenExpired(t *testing.T) {
	service, refreshTokenRepository, _ := newServiceAndMocks(t)

	refreshToken := refreshToken
	refreshToken.ExpirationTime = time.Now().Add(-time.Second)
	refreshTokenRepository.
		On("GetByID", refreshToken.ID).
		Return(&refreshToken, nil)

	var unauthorizedError *domain.UnauthorizedError
	_, err := service.RefreshSession(session, userIP)
	assert.ErrorAs(t, err, &unauthorizedError)
}

func TestRefreshSession_NewIP(t *testing.T) {
	service, refreshTokenRepository, emailService := newServiceAndMocks(t)

	refreshTokenRepository.
		On("GetByID", refreshToken.ID).
		Return(&refreshToken, nil)
	refreshTokenRepository.
		On("Create", mock.AnythingOfType("*domain.RefreshToken")).
		Return(uuid.New(), nil)
	refreshTokenRepository.
		On("DeleteByID", refreshToken.ID).
		Return(nil)
	emailService.
		On("SendSupportEmailToUser", userID, mock.Anything).
		Return(nil)

	_, err := service.RefreshSession(session, userIP+"1")
	assert.NoError(t, err)
}

func newServiceAndMocks(t *testing.T) (*AuthService, *mocks.RefreshTokenRepository, *mocks.EmailService) {
	refreshTokenRepository := mocks.NewRefreshTokenRepository(t)
	emailService := mocks.NewEmailService(t)
	service := NewAuthService(
		refreshTokenRepository,
		emailService,
		jwtPrivateKey,
		accessTokenDuration,
		refreshTokenDuration,
	)

	return service, refreshTokenRepository, emailService
}

func MustGenerateBcryptHashFromPassword(password []byte, cost int) []byte {
	hash, err := bcrypt.GenerateFromPassword(password, cost)
	if err != nil {
		panic(err)
	}
	return hash
}

func mustSignJWTString(jwt *jwt.Token, privateKey []byte) string {
	signedJWTStr, err := jwt.SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	return signedJWTStr
}
