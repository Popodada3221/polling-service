package domain

import (
	"errors"
	"time"
)

type Poll struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Options   []Option   `json:"options"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Option struct {
	ID         string `json:"id"`
	PollID     string `json:"poll_id"`
	Text       string `json:"text"`
	VotesCount int64  `json:"votes_count"`
}

type Vote struct {
	ID        string    `json:"id"`
	PollID    string    `json:"poll_id"`
	OptionID  string    `json:"option_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type VoteRequest struct {
	OptionID string `json:"option_id"`
	UserID   string `json:"user_id"`
}

type CreatePollRequest struct {
	Title     string     `json:"title" validate:"required,min=3,max=200"`
	Options   []string   `json:"options" validate:"required,min=2,max=10"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
type PollResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Options   []Option   `json:"options"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
type OptionResponse struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	VotesCount int64  `json:"votes_count"`
}

var ErrAlreadyVoted = errors.New("user already voted in this poll")
var ErrInvalidOption = errors.New("option is invalid")
