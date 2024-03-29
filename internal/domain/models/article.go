package models

import "time"

type Article struct {
	ID          int        `json:"id,omitempty"`
	Title       string     `json:"title,omitempty"`
	Content     string     `json:"content,omitempty"`
	PublishDate *time.Time `json:"publish_date,omitempty"`
	AuthorID    int        `json:"author_id,omitempty"`
}
