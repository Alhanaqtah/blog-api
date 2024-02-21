package article

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"blog-api/internal/domain/models"
	resp "blog-api/internal/lib/api/response"
	"blog-api/internal/lib/jwt"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/service/article"
	"blog-api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Service interface {
	GetAll() ([]models.Article, error)
	GetByID(id int) (*models.Article, error)
	Create(art *models.Article) error
	Update(art *models.Article) error
	Remove(id int) error
}

type Article struct {
	log     *slog.Logger
	service Service
	secret  string
}

func New(log *slog.Logger, service Service, secret string) *Article {
	return &Article{
		log:     log,
		service: service,
		secret:  secret,
	}
}

func (a *Article) Register() func(r chi.Router) {
	return func(r chi.Router) {
		// Public routes
		r.Get("/", a.getAll)
		r.Get("/{id}", a.getByID)

		// Require auth
		r.Group(func(r chi.Router) {
			tokenAuth := jwtauth.New("HS256", []byte(a.secret), nil)
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Post("/", a.create)
			r.Put("/{id}", a.update)
			r.Delete("/{id}", a.remove)
		})
	}
}

func (a *Article) getAll(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.article.getAll"

	log := a.log.With(slog.String("op", op))

	// Send to service layer
	articles, err := a.service.GetAll()
	if err != nil {
		log.Error("failed to get all articles", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status:   resp.StatusOk,
		Articles: &articles,
	})
}

func (a *Article) create(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.article.create"

	log := a.log.With(slog.String("op", op))

	var art models.Article
	err := render.DecodeJSON(r.Body, &art)
	if err != nil {
		log.Error("failed to decode request", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	satisfied, err := jwt.CheckClaim(r.Context(), "uid", strconv.Itoa(art.UserID))
	if err != nil {
		log.Error("failed to check permission", slog.String("user_id", strconv.Itoa(art.UserID)), sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}
	if !satisfied {
		log.Debug("user doesn't have permission", slog.Int("user_id", art.UserID))
		render.JSON(w, r, resp.Err("not enough rights"))
		return
	}

	// Validation
	if art.Title == "" {
		log.Debug("failed to create article: title is empty")
		render.JSON(w, r, resp.Err("title is empty"))
		return
	}
	if art.Content == "" {
		log.Debug("failed to create article: content is empty")
		render.JSON(w, r, resp.Err("content is empty"))
		return
	}

	// Send to service layer
	err = a.service.Create(&art)
	if err != nil {
		log.Error("failed to create article", sl.Error(err))
		if errors.Is(err, article.ErrArticleExists) {
			render.JSON(w, r, resp.Err("article title already taken"))
			return
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (a *Article) getByID(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.article.getByID"

	log := a.log.With(slog.String("op", op))

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to get \"id\" url param", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Send to service layer
	artcl, err := a.service.GetByID(id)
	if err != nil {
		log.Error("failed to get article by id", sl.Error(err))
		if errors.Is(err, article.ErrArticleNotFound) {
			render.JSON(w, r, resp.Err("article not found"))
			return
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	var art []models.Article
	art = append(art, *artcl)

	// Write to response
	render.JSON(w, r, resp.Response{
		Status:   resp.StatusOk,
		Articles: &art,
	})
}

func (a *Article) update(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.article.update"

	log := a.log.With(slog.String("op", op))

	articleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to get \"id\" url param", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	var art models.Article
	err = render.DecodeJSON(r.Body, &art)
	if err != nil {
		log.Error("failed to decode request", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Send to service layer
	ar, err := a.service.GetByID(articleID)
	if err != nil {
		log.Error("failed to get article by id", sl.Error(err))
		if errors.Is(err, storage.ErrArticleNotFound) {
			render.JSON(w, r, resp.Err("article not found"))
			return
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	satisfied, err := jwt.CheckClaim(r.Context(), "uid", strconv.Itoa(ar.UserID))
	if err != nil {
		log.Error("failed to check permission")
		render.JSON(w, r, resp.Err("internal error"))
		return
	}
	if !satisfied {
		log.Error("user doesn't have permission")
		render.JSON(w, r, resp.Err("not enough rights"))
		return
	}

	// Send to service layer
	err = a.service.Update(&art)
	if err != nil {
		log.Error("failed to update article", sl.Error(err))
		if errors.Is(err, article.ErrArticleNotFound) {
			render.JSON(w, r, resp.Err("article not found"))
			return
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (a *Article) remove(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.article.remove"

	log := a.log.With(slog.String("op", op))

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to get \"id\" url param", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Send to service layer
	art, err := a.service.GetByID(id)
	if err != nil {
		log.Error("failed to get user by id", sl.Error(err))
		if errors.Is(err, article.ErrArticleNotFound) {
			render.JSON(w, r, resp.Err("article not found"))
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	satisfied, err := jwt.CheckClaim(r.Context(), "uid", strconv.Itoa(art.UserID))
	if err != nil {
		log.Error("failed to check permission")
		render.JSON(w, r, resp.Err("internal error"))
		return
	}
	if !satisfied {
		log.Error("user doesn't have permission")
		render.JSON(w, r, resp.Err("not enough rights"))
		return
	}

	// Send to service layer
	err = a.service.Remove(id)
	if err != nil {
		log.Error("failed to remove article", sl.Error(err))
		if errors.Is(err, article.ErrArticleNotFound) {
			render.JSON(w, r, resp.Err("article not found"))
			return
		}
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}
