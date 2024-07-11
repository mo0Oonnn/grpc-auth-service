package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"github.com/mo0Oonnn/sso/internal/models"
	"github.com/mo0Oonnn/sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const operation = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context, email string, passwordHash []byte) (int64, error) {
	const operation = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	result, err := stmt.ExecContext(ctx, email, passwordHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", operation, err)
		}

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	return id, nil
}

// User returns user by email.
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const operation = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT * from users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", operation, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

// TODO: переделать проверку на отдельную таблицу, а не на поле
//
// IsAdmin checks if the user is an admin.
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operation = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", operation, storage.ErrNotFound)
		}
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return isAdmin, nil
}

// App returns app by id.
func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const operation = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT * from apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", operation, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}

	return app, nil
}
