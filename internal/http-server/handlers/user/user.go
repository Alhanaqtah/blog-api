package user

import (
	"blog-api/internal/domain/models"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	req "blog-api/internal/lib/api/request"
	resp "blog-api/internal/lib/api/response"
	"blog-api/internal/lib/logger/sl"
	"blog-api/internal/service/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Service interface {
	Remove(id int64) error
	UserByID(id int64) (models.User, error)
	Register(username, password string) error
	Login(username, password, secret string) (token string, err error)
	UpdateUserName(id int64, userName string) error
	UpdateStatus(id int64, status string) error
}

type User struct {
	log     *slog.Logger
	service Service
	secret  string
}

func New(log *slog.Logger, service Service, secret string) *User {
	return &User{
		log:     log,
		service: service,
		secret:  secret,
	}
}

func (u *User) Register() func(r chi.Router) {
	return func(r chi.Router) {
		// Public routes
		r.Get("/{id}", u.get)
		r.Post("/login", u.login)
		r.Post("/register", u.register)

		// Require auth
		r.Group(func(r chi.Router) {
			tokenAuth := jwtauth.New("HS256", []byte(u.secret), nil)
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Put("/{id}", u.update)
			r.Delete("/{id}", u.remove)
		})
	}
}

func (u *User) login(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.login"

	var cred req.Credentials
	err := render.DecodeJSON(r.Body, &cred)
	if err != nil {
		u.log.Error("%s: %w", op, err)
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	// Validate user creds
	if cred.Username == "" {
		u.log.Debug("username is empty", slog.String("op", op))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "invalid credentials: username is empty",
		})
		return
	}

	if cred.Password == "" {
		u.log.Debug("password is empty", slog.String("op", op))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "invalid credentials: password is empty",
		})
		return
	}

	token, err := u.service.Login(cred.Username, cred.Password, u.secret)
	if err != nil {
		u.log.Debug("can't create token", sl.Error(err))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
		Token:  token,
	})
}

func (u *User) register(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.register"

	var cred req.Credentials
	err := render.DecodeJSON(r.Body, &cred)
	if err != nil {
		u.log.Error("%s: %w", op, err)
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	// Validate user creds
	if cred.Username == "" {
		u.log.Debug("username is empty", slog.String("op", op))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "invalid credentials: username is empty",
		})
		return
	}

	if cred.Password == "" {
		u.log.Debug("password is empty", slog.String("op", op))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "invalid credentials: password is empty",
		})
		return
	}

	// Send to service layer
	err = u.service.Register(cred.Username, cred.Password)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			u.log.Debug("user already exists", slog.String("op", op))
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "user already exists",
			})
			return
		}

		u.log.Info("error registering new user", slog.String("op", op))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (u *User) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	user, err := u.service.UserByID(int64(id))
	if err != nil {
		u.log.Debug("can't get user by id", sl.Error(err))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
		User:   &user,
	})
}

func (u *User) update(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return
	}

	c := claims["uid"]
	uid, ok := c.(float64)
	if !ok {
		u.log.Debug("error getting uid")
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if id != int(uid) {
		u.log.Debug("user don't have permission")
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "there are not enough necessary rights",
		})
		return
	}

	var status req.Status
	err = render.DecodeJSON(r.Body, &status)
	if err != nil {
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	// Validation
	if status.UserName != "" {
		err := u.service.UpdateUserName(int64(id), status.UserName)
		if err != nil {
			u.log.Debug("error updating username", sl.Error(err))
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "internal error",
			})
			return
		}
	}

	if status.Status != "" {
		err := u.service.UpdateStatus(int64(id), status.Status)
		if err != nil {
			u.log.Debug("error updating status", sl.Error(err))
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "internal error",
			})
			return
		}
	}

	// Response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (u *User) remove(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать систему ролей: пользватель, админ
	// TODO: делать токен недействитеьным после удаления

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return
	}

	c := claims["uid"]
	uid, ok := c.(float64)
	if !ok {
		u.log.Debug("error getting uid")
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if id != int(uid) {
		u.log.Debug("user don't have permission")
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "there are not enough necessary rights",
		})
		return
	}

	err = u.service.Remove(int64(id))
	if err != nil {
		u.log.Debug("can't remove user", sl.Error(err))
		render.JSON(w, r, resp.Response{
			Status: resp.StatusError,
			Error:  "internal error",
		})
		return
	}

	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}
