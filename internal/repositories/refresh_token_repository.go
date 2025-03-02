package repositories

import (
	"auth/internal/domain"
	authservice "auth/internal/domain/services/auth-service"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type RefreshTokenRepository struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (s *RefreshTokenRepository) Create(token *domain.RefreshToken) (uuid.UUID, error) {
	query, args, err := s.builder.
		Insert("refresh_tokens").
		Columns(`value_hash, expires_at`).
		Values(token.ValueHash, token.ExpirationTime).
		Suffix("RETURNING \"id\"").
		ToSql()
	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "build query")
	}

	var id uuid.UUID
	err = s.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "execute query")
	}

	return id, nil
}

func (s *RefreshTokenRepository) GetByID(id uuid.UUID) (*domain.RefreshToken, error) {
	query, args, err := s.builder.
		Select("id, value_hash, expires_at").
		From("refresh_tokens").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	var refreshToken domain.RefreshToken
	err = s.db.QueryRow(query, args...).Scan(
		&refreshToken.ID, &refreshToken.ValueHash, &refreshToken.ExpirationTime,
	)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	return &refreshToken, nil
}

func (s *RefreshTokenRepository) DeleteByID(id uuid.UUID) error {
	query, args, err := s.builder.
		Delete("refresh_tokens").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "build query")
	}

	err = s.db.QueryRow(query, args...).Err()
	if err != nil {
		return errors.Wrap(err, "execute query")
	}

	return nil
}

func (s *RefreshTokenRepository) DeleteAllExpired() error {
	query, args, err := s.builder.
		Delete("refresh_tokens").
		Where("NOW() > expires_at").
		ToSql()
	if err != nil {
		return errors.Wrap(err, "build query")
	}

	err = s.db.QueryRow(query, args...).Err()
	if err != nil {
		return errors.Wrap(err, "execute query")
	}

	return nil
}

var _ authservice.RefreshTokenRepository = &RefreshTokenRepository{}
