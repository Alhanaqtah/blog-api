package article

import (
	"github.com/go-chi/jwtauth/v5"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Article struct {
	log    *slog.Logger
	secret string
}

func New(log *slog.Logger, secret string) *Article {
	return &Article{
		log:    log,
		secret: secret,
	}
}

func (a *Article) Register() func(r chi.Router) {
	return func(r chi.Router) {
		// Public routes
		r.Get("/", a.getAllArticles)
		r.Get("/{id}", a.getArticleByID)

		// Require auth
		r.Group(func(r chi.Router) {
			tokenAuth := jwtauth.New("HS256", []byte(a.secret), nil)
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Post("/", a.createArticle)
			r.Put("/{id}", a.correctArticle)
			r.Delete("/{id}", a.removeArticle)
		})
	}
}

func (a *Article) getAllArticles(w http.ResponseWriter, r *http.Request) {

}

func (a *Article) createArticle(w http.ResponseWriter, r *http.Request) {

}

func (a *Article) getArticleByID(w http.ResponseWriter, r *http.Request) {

}

func (a *Article) correctArticle(w http.ResponseWriter, r *http.Request) {

}

func (a *Article) removeArticle(w http.ResponseWriter, r *http.Request) {

}
