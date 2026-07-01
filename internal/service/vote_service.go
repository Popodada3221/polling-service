package service

import (
	"context"
	"fmt"
	"log/slog"
	"polling-service/internal/domain"
	"polling-service/internal/repository"
)

type VoteService struct {
	voteRepo *repository.VoteRepository
	pollRepo *repository.PollRepository
	cache    *repository.CacheRepository
}

func NewVoteService(pollRepo *repository.PollRepository, voteRepo *repository.VoteRepository, cache *repository.CacheRepository) *VoteService {
	return &VoteService{
		pollRepo: pollRepo,
		voteRepo: voteRepo,
		cache:    cache,
	}

}
func (s *VoteService) Vote(ctx context.Context, pollID, optionID, userID string) (*domain.Poll, error) {
	poll, err := s.pollRepo.GetPoll(ctx, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get poll: %w", err)
	}
	if poll == nil {
		return nil, ErrPollNotFound
	}
	if poll.DeletedAt != nil {
		return nil, ErrPollNotFound
	}

	var found bool
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			found = true
			break
		}
	}
	if !found {
		return nil, domain.ErrInvalidOption
	}

	vote := &domain.Vote{
		PollID:   pollID,
		OptionID: optionID,
		UserID:   userID,
	}

	err = s.voteRepo.CreateVote(ctx, vote)
	if err != nil {
		return nil, err
	}

	if err := s.cache.DeletePoll(ctx, pollID); err != nil {
		slog.Warn("Failed to invalidate cache", "poll_id", pollID, "error", err)
	}

	updatedPoll, err := s.pollRepo.GetPoll(ctx, pollID)

	if err != nil {
		return nil, err
	}

	//update cache

	if err := s.cache.SetPoll(ctx, updatedPoll); err != nil {
		slog.Warn("Failed to update cache", "poll_id", pollID, "error", err)
	}

	return updatedPoll, nil

}

func (s *VoteService) ListVotes(ctx context.Context, page, pageSize int) ([]domain.Vote, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.voteRepo.ListVotes(ctx, pageSize, offset)
}
