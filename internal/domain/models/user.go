package models

import "time"

type User struct {
	ID               int64      `json:"id,omitempty"`
	RegistrationDate *time.Time `json:"registration_date,omitempty"`
	Status           string     `json:"status,omitempty"`
	ArticlesID       []int64    `json:"articles_id,omitempty"`
	Credentials      `json:"credentials,omitempty"`
}

type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	PassHash []byte `json:"pass_hash,omitempty"`
}
