package repository

import (
	"context"
	"fmt"
	"polling-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PollRepository struct {
	db *pgxpool.Pool
}

func NewPollRepository(db *pgxpool.Pool) *PollRepository {

	return &PollRepository{
		db: db,
	}
}

func (r *PollRepository) CreatePoll(ctx context.Context, poll *domain.Poll, options []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	pollID := uuid.New().String()
	poll.ID = pollID

	query := `INSERT INTO polls (id, title, expires_at) VALUES ($1, $2, $3) RETURNING created_at`
	err = tx.QueryRow(ctx, query, pollID, poll.Title, poll.ExpiresAt).Scan(&poll.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create poll: %w", err)
	}

	for _, optionText := range options {
		optionID := uuid.New().String()
		query = `INSERT INTO options (id, poll_id, text) VALUES ($1, $2, $3)`

		_, err = tx.Exec(ctx, query, optionID, pollID, optionText)
		if err != nil {
			return fmt.Errorf("failed to create option: %w", err)
		}
		poll.Options = append(poll.Options, domain.Option{
			ID:     optionID,
			PollID: pollID,
			Text:   optionText,
		})
	}
	return tx.Commit(ctx)
}

func (r *PollRepository) GetPoll(ctx context.Context, id string) (*domain.Poll, error) {
	query := `SELECT id, title, created_at, expires_at FROM polls WHERE id = $1 AND deleted_at IS NULL`
	poll := &domain.Poll{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&poll.ID,
		&poll.Title,
		&poll.CreatedAt,
		&poll.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get poll by ID: %w", err)
	}

	query = `SELECT id, poll_id, text, votes_count FROM options WHERE poll_id = $1`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get options for poll: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var opt domain.Option
		err := rows.Scan(&opt.ID, &opt.PollID, &opt.Text, &opt.VotesCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		poll.Options = append(poll.Options, opt)
	}

	return poll, nil
}

func (r *PollRepository) ListPolls(ctx context.Context, limit, offset int) ([]domain.Poll, error) {
	query := `SELECT id, title, created_at, expires_at FROM polls WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list polls: %w", err)
	}
	defer rows.Close()

	var polls []domain.Poll
	for rows.Next() {
		var p domain.Poll
		err := rows.Scan(&p.ID, &p.Title, &p.CreatedAt, &p.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan poll: %w", err)
		}
		polls = append(polls, p)
	}

	return polls, nil
}
