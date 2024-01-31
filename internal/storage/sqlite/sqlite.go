package sqlite

import (
	"blog-api/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"time"

	"blog-api/internal/storage"

	_ "github.com/mattn/go-sqlite3"
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
			status TEXT
		);
		
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			publish_date DATETIME NOT NULL,
			author_id INTEGER REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS user_articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id),
			article_id INTEGER REFERENCES articles(id)
		);
`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Register(ctx context.Context, username string, passHash []byte) error {
	const op = "storage.sqlite.Register"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (name, pass_hash, registration_date) VALUES (?, ?, ?)`)
	if err != nil {
		return err
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

func (s *Storage) User(ctx context.Context, username string) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.PrepareContext(ctx, `SELECT id, name, pass_hash FROM users WHERE name = ?`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res := stmt.QueryRowContext(ctx, username)

	var user models.User
	err = res.Scan(&user.ID, &user.Username, &user.PassHash)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
