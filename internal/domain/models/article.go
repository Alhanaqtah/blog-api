package models

import "time"

type Article struct {
	ID          int         `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	Content     string      `json:"content,omitempty"`
	PublishDate *time.Timer `json:"publish_date,omitempty"`
	UserID      int         `json:"user_id,omitempty"`
}
