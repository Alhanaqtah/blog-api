package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"blog-api/internal/domain/models"
	"blog-api/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			pass_hash BLOB NOT NULL,
			registration_date DATETIME NOT NULL,
			status TEXT DEFAULT ''
		);
		
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			publish_date DATETIME NOT NULL,
			author_id INTEGER REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS users_articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id),
			article_id INTEGER REFERENCES articles(id)
		);
`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Register(ctx context.Context, username string, passHash []byte) error {
	const op = "storage.sqlite.Register"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (name, pass_hash, registration_date) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, username, passHash, time.Now())
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UserByName(ctx context.Context, username string) (models.User, error) {
	const op = "storage.sqlite.UserByName"

	stmt, err := s.db.PrepareContext(ctx, `SELECT id, name, pass_hash FROM users WHERE name = ?`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res := stmt.QueryRowContext(ctx, username)

	var user models.User
	err = res.Scan(&user.ID, &user.Username, &user.PassHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UserByID(ctx context.Context, id int64) (models.User, error) {
	const op = "storage.sqlite.UserByID"

	stmt, err := s.db.PrepareContext(ctx, `SELECT id, name, registration_date, status FROM users WHERE id = ?`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res := stmt.QueryRowContext(ctx, id)

	var user models.User
	err = res.Scan(&user.ID, &user.Username, &user.RegistrationDate, &user.Status)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) Remove(ctx context.Context, id int64) error {
	const op = "storage.sqlite.Remove"

	stmt, err := s.db.PrepareContext(ctx, `DELETE FROM users WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateUserName(ctx context.Context, id int64, username string) error {
	const op = "storage.service.UpdateUserName"

	stmt, err := s.db.PrepareContext(ctx, `UPDATE users SET name = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, username, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateStatus(ctx context.Context, id int64, status string) error {
	const op = "storage.sqlite.UpdateStatus"

	stmt, err := s.db.PrepareContext(ctx, `UPDATE users SET status = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, status, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
