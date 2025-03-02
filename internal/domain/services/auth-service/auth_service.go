package authservice

import (
	"auth/internal/domain"
	jwtutils "auth/internal/utils/jwt-utils"
	slogutils "auth/internal/utils/slog-utils"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const deleteExpiredTokensPeriod = time.Minute * 5

type AuthService struct {
	refreshTokenRepository RefreshTokenRepository
	emailService           EmailService
	accessTokenDuration    time.Duration
	refreshTokenDuration   time.Duration
	jwtPrivateKey          []byte
}

//go:generate mockery --name RefreshTokenRepository --filename refresh_token_repository.go
type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) (id uuid.UUID, err error)
	GetByID(id uuid.UUID) (*domain.RefreshToken, error)
	DeleteByID(id uuid.UUID) error
	DeleteAllExpired() error
}

//go:generate mockery --name EmailService --filename email_service.go
type EmailService interface {
	SendSupportEmailToUser(userID uuid.UUID, emailContent domain.EmailContent) error
}

func NewAuthService(
	refershTokenRepository RefreshTokenRepository,
	emailService EmailService,
	jwtPrivateKey []byte,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {

	go func() {
		timer := time.NewTicker(deleteExpiredTokensPeriod)
		for {
			<-timer.C
			err := refershTokenRepository.DeleteAllExpired()
			if err != nil {
				slogutils.Error("delete all expired refresh tokens", err)
			}
		}
	}()

	return &AuthService{
		refreshTokenRepository: refershTokenRepository,
		emailService:           emailService,
		accessTokenDuration:    accessTokenDuration,
		refreshTokenDuration:   refreshTokenDuration,
		jwtPrivateKey:          jwtPrivateKey,
	}
}

func (s *AuthService) CreateSession(
	userID uuid.UUID, requestIP string,
) (*domain.Session, error) {

	refreshTokenValueBytes, err := generateRefreshTokenValueBytes()
	if err != nil {
		return nil, errors.Wrap(err, "generate refresh token bytes")
	}
	refreshTokenHash, err := bcrypt.
		GenerateFromPassword(refreshTokenValueBytes, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "hash refresh token")
	}

	refreshTokenExpTime := time.Now().Add(s.refreshTokenDuration)
	refreshToken := &domain.RefreshToken{
		ValueHash:      refreshTokenHash,
		ExpirationTime: refreshTokenExpTime,
	}
	refreshTokenID, err := s.refreshTokenRepository.Create(refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "save refresh token")
	}

	accessTokenExpTime := time.Now().Add(s.accessTokenDuration).Unix()
	accessTokenJWT := jwt.NewWithClaims(accessTokenJWTSigningMethod,
		jwt.MapClaims{
			UserIDJWTClaimName:         userID.String(),
			UserIPJWTClaimName:         requestIP,
			ExpirationTimeJWTClaimName: accessTokenExpTime,
			RefreshTokenIDJWTClaimName: refreshTokenID,
		})
	accessTokenStr, err := accessTokenJWT.SignedString(s.jwtPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "sign access token")
	}

	return &domain.Session{
		AccessTokenSigned: []byte(accessTokenStr),
		RefreshTokenValue: refreshTokenValueBytes,
	}, nil
}

func (s *AuthService) RefreshSession(
	session *domain.Session, requestIP string,
) (*domain.Session, error) {

	accessTokenJWT, err := jwtutils.
		ParseAndValidateJWTToken(
			session.AccessTokenSigned,
			s.jwtPrivateKey,
			accessTokenJWTSigningMethod.Name)
	if err != nil {
		return nil, &domain.UnauthorizedError{
			Message: fmt.Sprintf("parse access token: %s", err)}
	}

	accessToken, err := parseAccessTokenFromJWT(accessTokenJWT)
	if err != nil {
		return nil, &domain.UnauthorizedError{
			Message: fmt.Sprintf("parse access token: %s", err)}
	}

	refreshToken, err := s.refreshTokenRepository.
		GetByID(accessToken.RefreshTokenID)
	if err != nil {
		return nil, &domain.UnauthorizedError{
			Message: "refresh token not found"}
	}

	if time.Now().After(refreshToken.ExpirationTime) {
		return nil, &domain.UnauthorizedError{
			Message: "refresh token not found"}
	}

	if bcrypt.CompareHashAndPassword(
		refreshToken.ValueHash,
		session.RefreshTokenValue,
	) != nil {
		return nil, &domain.UnauthorizedError{
			Message: "refresh token is invalid"}
	}

	err = s.refreshTokenRepository.DeleteByID(refreshToken.ID)
	if err != nil {
		return nil, errors.Wrap(err, "delete refresh token")
	}

	if requestIP != accessToken.UserIP {
		go func() {
			err := s.emailService.SendSupportEmailToUser(
				accessToken.UserID,
				newEmailContentRefreshFromNewIPWarning(requestIP))
			if err != nil {
				slogutils.Error("send warning email(refresh from new ip) error", err)
			}
		}()
	}

	return s.CreateSession(
		accessToken.UserID,
		requestIP,
	)
}
