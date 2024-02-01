package response

import (
	"blog-api/internal/domain/models"
)

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

type Response struct {
	Status string       `json:"status"`
	Error  string       `json:"error,omitempty"`
	Token  string       `json:"token,omitempty"`
	User   *models.User `json:"user,omitempty"`
}
