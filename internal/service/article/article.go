package article

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"blog-api/internal/domain/models"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/storage"
)

var (
	ErrArticleExists   = errors.New("article already exists")
	ErrArticleNotFound = errors.New("article not found")
)

type Storage interface {
	GetAllArticles(ctx context.Context) ([]models.Article, error)
	GetArticleByID(ctx context.Context, id int) (*models.Article, error)
	CreateArticle(ctx context.Context, userID int, title, content string, publishDate time.Time) error
	UpdateArticleTitle(ctx context.Context, id int, title string) error
	UpdateArticleContent(ctx context.Context, id int, content string) error
	RemoveArticle(ctx context.Context, id int) error
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

func (s *Service) GetAll() ([]models.Article, error) {
	const op = "service.article.GetAll"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to storage layer
	arts, err := s.storage.GetAllArticles(ctx)
	if err != nil {
		log.Error("failed to get all articles", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return arts, nil
}

func (s *Service) GetByID(id int) (*models.Article, error) {
	const op = "service.article.GetByID"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to storage layer
	art, err := s.storage.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrArticleNotFound) {
			log.Error("article not found", sl.Error(err))
			return nil, fmt.Errorf("%s: %w", op, ErrArticleNotFound)
		}
		log.Error("failed to get article", sl.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return art, nil
}

func (s *Service) Create(art *models.Article) error {
	const op = "service.article.Create"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to storage layer
	err := s.storage.CreateArticle(ctx, art.AuthorID, art.Title, art.Content, time.Now())
	if err != nil {
		if errors.Is(err, storage.ErrArticleExists) {
			log.Error("article not found", sl.Error(err))
			return fmt.Errorf("%s: %w", op, ErrArticleExists)
		}
		log.Error("failed to get art", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Update(art *models.Article) error {
	const op = "service.article.Update"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to storage layer
	var err error
	if art.Title != "" {
		err = s.storage.UpdateArticleTitle(ctx, art.ID, art.Title)
	}
	if art.Content != "" {
		err = s.storage.UpdateArticleContent(ctx, art.ID, art.Content)
	}
	if err != nil {
		/* if errors.As(err, &storage.ErrArticleNotFound) {
			log.Error("article not found", sl.Error(err))
			return fmt.Errorf("%s: %w", op, ErrArticleNotFound)
		} */
		log.Error("failed to update article", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Remove(id int) error {
	const op = "service.article.RemoveUser"

	log := s.log.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send to storage layer
	err := s.storage.RemoveArticle(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrArticleNotFound) {
			log.Error("failed to remove article", sl.Error(err))
			return fmt.Errorf("%s: %w", op, ErrArticleNotFound)
		}
		log.Error("failed to remove article", sl.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
