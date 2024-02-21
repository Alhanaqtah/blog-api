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
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			pass_hash BLOB NOT NULL,
			registration_date DATETIME NOT NULL,
			status TEXT DEFAULT ''
		);
		
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			publish_date DATETIME NOT NULL,
			author_id INTEGER REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS users_articles (
			article_d INTEGER REFERENCES articles(id)
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

// ### User ### //

func (s *Storage) Register(ctx context.Context, username string, passHash []byte, regestrationDate time.Time) error {
	const op = "storage.sqlite.Register"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (name, pass_hash, registration_date) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, username, passHash, regestrationDate)
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

func (s *Storage) UserByID(ctx context.Context, id int) (models.User, error) {
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
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) RemoveUser(ctx context.Context, id int) error {
	const op = "storage.sqlite.RemoveUser"

	stmt, err := s.db.PrepareContext(ctx, `DELETE FROM users WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateUserName(ctx context.Context, id int, username string) error {
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

func (s *Storage) UpdateStatus(ctx context.Context, id int, status string) error {
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

// ### Article ### //

func (s *Storage) GetAllArticles(ctx context.Context) ([]models.Article, error) {
	const op = "storage.sqlite.GetAllArticles"

	stmt, err := s.db.PrepareContext(ctx, `SELECT * FROM articles`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var arts []models.Article
	for rows.Next() {
		var art models.Article

		err = rows.Scan(&art.ID, &art.Title, &art.Content, &art.PublishDate, &art.UserID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		arts = append(arts, art)
	}

	return arts, nil
}

func (s *Storage) GetArticleByID(ctx context.Context, id int) (*models.Article, error) {
	const op = "storage.sqlite.GetArticleByID"

	stmt, err := s.db.PrepareContext(ctx, `SELECT title, content, publish_date, user_id FROM articles WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var art models.Article
	err = row.Scan(&art.Title, &art.Content, &art.PublishDate, &art.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &art, nil
}

func (s *Storage) CreateArticle(ctx context.Context, userID int, title, content string, publishDate time.Time) error {
	const op = "storage.sqlite.CreateArticle"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO articles (title, content, publish_date, user_id) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, title, content, publishDate, userID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrArticleExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateArticleTitle(ctx context.Context, id int, title string) error {
	const op = "storage.sqlite.UpdateArticleTitle"

	stmt, err := s.db.PrepareContext(ctx, `UPDATE articles SET title = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, title, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateArticleContent(ctx context.Context, id int, content string) error {
	const op = "storage.sqlite.UpdateArticleContent"

	stmt, err := s.db.PrepareContext(ctx, `UPDATE articles SET content = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, content, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveArticle(ctx context.Context, id int) error {
	const op = "storage.sqlite.RemoveArticle"

	stmt, err := s.db.PrepareContext(ctx, `DELETE FROM articles WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
