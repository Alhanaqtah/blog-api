package user

import (
	"blog-api/internal/domain/models"
	"blog-api/internal/lib/jwt"
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
	Remove(id int) error
	UserByID(id int) (models.User, error)
	Register(userName, password string) error
	Login(userName, password, secret string) (token string, err error)
	UpdateUserName(id int, userName string) error
	UpdateStatus(id int, status string) error
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

	log := u.log.With(slog.String("op", op))

	var cred req.Credentials
	err := render.DecodeJSON(r.Body, &cred)
	if err != nil {
		log.Error("failed to decode request", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Validate user creds
	if cred.UserName == "" {
		u.log.Error("user name is empty")
		render.JSON(w, r, resp.Err("invalid credentials: user name is empty"))
		return
	}

	if cred.Password == "" {
		u.log.Error("password is empty")
		render.JSON(w, r, resp.Err("invalid credentials: password is empty"))
		return
	}

	// Send to service layer
	token, err := u.service.Login(cred.UserName, cred.Password, u.secret)
	if err != nil {
		u.log.Error("failed to create new token", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
		Token:  token,
	})
}

func (u *User) register(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.register"

	log := u.log.With(slog.String("op", op))

	var cred req.Credentials
	err := render.DecodeJSON(r.Body, &cred)
	if err != nil {
		log.Error("failed to decode request", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Validate user creds
	if cred.UserName == "" {
		u.log.Error("user name is empty")
		render.JSON(w, r, resp.Err("invalid credentials: user name is empty"))
		return
	}

	if cred.Password == "" {
		u.log.Error("password is empty")
		render.JSON(w, r, resp.Err("password is empty"))
		return
	}

	// Send to service layer
	err = u.service.Register(cred.UserName, cred.Password)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			u.log.Error("failed to register user", sl.Error(err))
			render.JSON(w, r, resp.Err("user already exists"))
			return
		}

		u.log.Info("failed to register new user", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (u *User) get(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.get"

	log := u.log.With(slog.String("op", op))

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to get \"id\" url param", sl.Error(err))
	}

	// Send to service layer
	user, err := u.service.UserByID(id)
	if err != nil {
		u.log.Error("failed to get user by id", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
		User:   &user,
	})
}

func (u *User) update(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.update"

	log := u.log.With(slog.String("op", op))

	// Getting id from url params
	id := chi.URLParam(r, "id")

	// Checking user permission
	satisfied, err := jwt.CheckClaim(r.Context(), "uid", id)
	if !satisfied {
		log.Error("user doesn't have permission", slog.String("user_id", id))
		render.JSON(w, r, resp.Err("not enough rights"))
		return
	}
	if err != nil {
		log.Error("failed to check permission", slog.String("user_id", id))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	//// Checking user permission
	//_, claims, err := jwtauth.FromContext(r.Context())
	//if err != nil {
	//	log.Error("failed to get token claims", sl.Error(err))
	//	render.JSON(w, r, resp.Err("internal error"))
	//}
	//
	////// Getting uid from token claims
	//c := claims["uid"]
	//uid, ok := c.(float64)
	//if !ok {
	//	log.Error("failed to get uid from claims", sl.Error(err))
	//	render.JSON(w, r, resp.Err("internal error"))
	//}

	////// checking if user have permission
	//if id != int(uid) {
	//	log.Debug("user doesn't have permission", slog.Int("user_id", id))
	//	render.JSON(w, r, resp.Err("not enough rights"))
	//	return
	//}

	var upd req.Update
	err = render.DecodeJSON(r.Body, &upd)
	if err != nil {
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Error("failed to convert str to int", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Validation
	if upd.UserName != "" {
		// Send to service layer

		err := u.service.UpdateUserName(userID, upd.UserName)
		if err != nil {
			u.log.Error("failed to update user name", sl.Error(err))
			render.JSON(w, r, resp.Err("internal error"))
			return
		}
	}

	err = u.service.UpdateStatus(userID, upd.Status)
	if err != nil {
		u.log.Error("failed to update user status", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}

func (u *User) remove(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать систему ролей: пользватель, админ
	// TODO: делать токен недействитеьным после удаления
	const op = "handlers.user.remove"

	log := u.log.With(slog.String("op", op))

	// Checking user permission
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error("failed to get token claims", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
	}

	//// Getting uid from token claims
	c := claims["uid"]
	uid, ok := c.(float64)
	if !ok {
		log.Error("failed to get uid from claims", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
	}

	//// Getting id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to get \"id\" url param", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
	}

	//// checking if user have permission
	if id != int(uid) {
		log.Debug("user doesn't have permission", slog.Int("user_id", id))
		render.JSON(w, r, resp.Err("not enough rights"))
		return
	}

	// Send to service layer
	err = u.service.Remove(id)
	if err != nil {
		u.log.Error("failed to remove user", sl.Error(err))
		render.JSON(w, r, resp.Err("internal error"))
		return
	}

	// Write to response
	render.JSON(w, r, resp.Response{
		Status: resp.StatusOk,
	})
}
