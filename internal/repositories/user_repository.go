package repositories

import "github.com/google/uuid"

type UserEmailsRepositoryMock struct{}

func (s *UserEmailsRepositoryMock) GetUserEmail(userID uuid.UUID) (string, error) {
	return "user@gmail.com", nil
}

func NewUserRepositoryMock() *UserEmailsRepositoryMock {
	return &UserEmailsRepositoryMock{}
}
