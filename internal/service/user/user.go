package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"log/slog"
	"time"

	"blog-api/internal/domain/models"
	"blog-api/internal/lib/jwt"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("username already taken")
	ErrUserNotFound = errors.New("user not found")
)

type Storage interface {
	//Correct(context context.Context, user models.User) error
	//Remove(context context.Context, id int64) error
	Register(ctx context.Context, username string, passHash []byte) error
	UserByName(ctx context.Context, username string) (models.User, error)
	UserByID(ctx context.Context, id int64) (models.User, error)
}

type Service struct {
	log      *slog.Logger
	storage  Storage
	tokenTTL time.Duration
}

func New(log *slog.Logger, storage Storage, ttl time.Duration) *Service {
	return &Service{
		log:      log,
		storage:  storage,
		tokenTTL: ttl,
	}
}

func (u *Service) Register(username string, password string) error {
	const op = "service.user.Register"

	log := u.log.With(slog.String("op", op))

	// Hashing password
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Debug("error generating hash from password", sl.Error(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = u.storage.Register(ctx, username, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Debug("user already exists", sl.Error(err))
			return fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Debug("error while registering user", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *Service) Login(username string, password string, secret string) (token string, err error) {
	const op = "service.user.Login"

	log := u.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := u.storage.UserByName(ctx, username)
	if err != nil {
		if errors.As(err, storage.ErrUserNotFound) {
			log.Debug("user not found", sl.Error(err))
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Debug("incorrect password")
		return "", fmt.Errorf("%s: incorrect password: %w", op, err)
	}

	token, err = jwt.NewToken(user, u.tokenTTL, secret)
	if err != nil {
		log.Debug("failed to create token", sl.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *Service) UserByID(id int64) (models.User, error) {
	const op = "service.user.UserByID"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := s.storage.UserByID(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrNotFound {
			log.Debug("user not found", sl.Error(sqliteErr))
		}
		log.Debug("failed get user", sl.Error(err))
		return models.User{}, err
	}

	return user, nil
}
