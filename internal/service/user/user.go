package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"blog-api/internal/domain/models"
	"blog-api/internal/lib/jwt"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/storage"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("user name already taken")
	ErrUserNotFound = errors.New("user not found")
)

type Storage interface {
	Remove(ctx context.Context, id int) error
	UpdateUserName(ctx context.Context, id int, userName string) error
	UpdateStatus(ctx context.Context, id int, status string) error
	UserByID(ctx context.Context, id int) (models.User, error)
	UserByName(ctx context.Context, userName string) (models.User, error)
	Register(ctx context.Context, userName string, passHash []byte) error
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

func (s *Service) Register(userName, password string) error {
	const op = "service.user.Register"

	log := s.log.With(slog.String("op", op))

	// Hashing password
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash from password", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to data layer
	err = s.storage.Register(ctx, userName, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("failed to register user", sl.Error(ErrUserExists))
			return fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to register user", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Login(userName, password, secret string) (token string, err error) {
	const op = "service.user.Login"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to data layer
	user, err := s.storage.UserByName(ctx, userName)
	if err != nil {
		if errors.As(err, &storage.ErrUserNotFound) {
			log.Error("failed to get user by name", sl.Error(ErrUserNotFound))
			return "", fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Checking if password correct
	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Error("incorrect password", sl.Error(err))
		return "", fmt.Errorf("%s: incorrect password: %w", op, err)
	}

	// Generating token
	token, err = jwt.NewToken(user, s.tokenTTL, secret)
	if err != nil {
		log.Error("failed to create new token", sl.Error(err))
		return "", fmt.Errorf("%s: failed to create new token: %w", op, err)
	}

	return token, nil
}

func (s *Service) UserByID(id int) (models.User, error) {
	const op = "service.user.UserByID"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to data layer
	user, err := s.storage.UserByID(ctx, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrNotFound {
			log.Error("user not found", ErrUserNotFound)
		}
		log.Error("failed get user", sl.Error(err))
		return models.User{}, err
	}

	return user, nil
}

func (s *Service) Remove(id int) error {
	const op = "service.user.Remove"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to data layer
	err := s.storage.Remove(ctx, id)
	if err != nil {
		log.Error("failed to remove user", sl.Error(err))
		return err
	}

	return nil
}

func (s *Service) UpdateUserName(id int, userName string) error {
	const op = "service.user.UpdateUserName"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to data layer
	err := s.storage.UpdateUserName(ctx, id, userName)
	if err != nil {
		log.Error("failed to update user name", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UpdateStatus(id int, userName string) error {
	const op = "service.user.UpdateStatus"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.storage.UpdateStatus(ctx, id, userName)
	if err != nil {
		log.Error("failed to update status", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
