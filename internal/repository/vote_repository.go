package repository

import (
	"context"
	"errors"
	"fmt"
	"polling-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VoteRepository struct {
	db *pgxpool.Pool
}

func NewVoteRepository(db *pgxpool.Pool) *VoteRepository {

	return &VoteRepository{
		db: db,
	}
}

func (r *VoteRepository) CreateVote(ctx context.Context, vote *domain.Vote) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	voteID := uuid.New().String()
	vote.ID = voteID

	query := `INSERT INTO votes (id, poll_id, option_id, user_id) VALUES ($1, $2, $3, $4) RETURNING created_at`
	err = tx.QueryRow(ctx, query, voteID, vote.PollID, vote.OptionID, vote.UserID).Scan(&vote.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyVoted
		}
		return fmt.Errorf("failed to create vote: %w", err)

	}
	query = `UPDATE options SET votes_count = votes_count + 1 WHERE id = $1`
	_, err = tx.Exec(ctx, query, vote.OptionID)
	if err != nil {
		return fmt.Errorf("failed to increment votes: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *VoteRepository) ListVotes(ctx context.Context, limit, offset int) ([]domain.Vote, error) {
	query := `SELECT id, poll_id, option_id, user_id, created_at FROM votes LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list votes: %w", err)
	}
	var votes []domain.Vote
	for rows.Next() {
		var v domain.Vote
		err := rows.Scan(&v.ID, &v.PollID, &v.OptionID, &v.UserID, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan poll: %w", err)
		}
		votes = append(votes, v)
	}

	return votes, nil

}
