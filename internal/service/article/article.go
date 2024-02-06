package article

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"blog-api/internal/domain/models"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/storage"
)

type Storage interface {
	GetAllUsers() (*[]models.Article, error)
	GetByID(id int) (*models.Article, error)
	Create(articleID, userID int, title, content string, publishDate time.Time) error
	UpdateTitle(id int, title string) error
	UpdateContent(id int, content string) error
	Remove(id int) error
}

type Service struct {
	log     *slog.Logger
	storage Storage
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s *Service) GetAll() (*[]models.Article, error) {
	const op = "service.article.GetAll"

	log := s.log.With(slog.String("op", op))

	// Send to storage layer
	arts, err := s.storage.GetAllUsers()
	if err != nil {
		log.Error("failed to get all users", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return arts, nil
}

func (s *Service) GetByID(id int) (*models.Article, error) {
	const op = "service.article.GetByID"

	log := s.log.With(slog.String("op", op))

	// Send to storage layer
	art, err := s.storage.GetByID(id)
	if err != nil {
		if errors.As(err, &storage.ErrArticleNotFound) {
			log.Error("article not found", sl.Error(err))
			return nil, fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		log.Error("failed to get article", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return art, nil
}

func (s *Service) Create(art *models.Article) error {
	const op = "service.article.GetByID"

	log := s.log.With(slog.String("op", op))

	// Send to storage layer
	err := s.storage.Create(art.ID, art.UserID, art.Title, art.Content, time.Now())
	if err != nil {
		if errors.As(err, &storage.ErrArticleNotFound) {
			log.Error("art not found", sl.Error(err))
			return fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		log.Error("failed to get art", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Update(art *models.Article) error {
	const op = "service.article.Update"

	log := s.log.With(slog.String("op", op))

	// Send to storage layer
	var err error
	if art.Title == "" {
		s.storage.UpdateTitle(art.ID, art.Title)
	}
	if art.Content == "" {
		s.storage.UpdateContent(art.ID, art.Content)
	}
	if err != nil {
		if errors.As(err, &storage.ErrArticleNotFound) {
			log.Error("article not found", sl.Error(err))
			return fmt.Errorf("%s: %w", op, storage.ErrArticleNotFound)
		}
		log.Error("failed to update article", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Remove(id int) error {
	const op = "service.article.Remove"

	log := s.log.With(slog.String("op", op))

	// Send to storage layer
	err := s.storage.Remove(id)
	if err != nil {
		log.Error("failed to remove article", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
