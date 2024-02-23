package storage

import "errors"

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	ErrArticleExists   = errors.New("article already exists")
	ErrArticleNotFound = errors.New("article not found")

	ErrUserNameTaken = errors.New("user name already taken")
	ErrTitleTaken    = errors.New("article title already taken")
)
