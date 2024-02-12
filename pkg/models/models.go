package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

type Snippet struct {
	ID      int       `json:"id,omitempty" :"id"`
	Title   string    `json:"title,omitempty" :"title"`
	Content string    `json:"content,omitempty" :"content"`
	Created time.Time `json:"created" :"created"`
	Expires time.Time `json:"expires" :"expires"`
}

type Department struct {
	ID            int    `json:"id,omitempty" :"id"`
	DepName       string `json:"dep___name,omitempty" :"dep___name"`
	StaffQuantity int    `json:"staff___quantity,omitempty" :"staff___quantity"`
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}
