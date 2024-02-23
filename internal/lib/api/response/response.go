package response

import (
	"blog-api/internal/domain/models"
)

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

type Response struct {
	Status   string            `json:"status"`
	Error    string            `json:"error,omitempty"`
	Token    string            `json:"token,omitempty"`
	Users    *[]models.User    `json:"users,omitempty"`
	Articles *[]models.Article `json:"articles,omitempty"`
}

func Err(errMsg string) Response {
	return Response{
		Status: StatusError,
		Error:  errMsg,
	}
}
