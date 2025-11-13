package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/model"
)

type UserRepository interface {
	GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	query := `
		SELECT user_id, firebase_uid, email, created_at, updated_at
		FROM users
		WHERE firebase_uid = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (user_id, firebase_uid, email, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query, user.UserID, user.FirebaseUID, user.Email)
	return err
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	query := `
		SELECT user_id, firebase_uid, email, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}