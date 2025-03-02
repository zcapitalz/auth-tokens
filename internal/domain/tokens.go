package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccessToken struct {
	UserID         uuid.UUID
	RefreshTokenID uuid.UUID
	UserIP         string
	ExpTime        time.Time
}

type RefreshToken struct {
	ID             uuid.UUID
	ValueHash      []byte
	ExpirationTime time.Time
}
