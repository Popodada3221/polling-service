package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"polling-service/internal/domain"
	"polling-service/internal/repository"
)

var (
	ErrPollNotFound = errors.New("poll not found")
	ErrInvalidPoll  = errors.New("invalid poll data")
)

type PollService struct {
	repo  *repository.PollRepository
	cache *repository.CacheRepository
}

func NewPollService(repo *repository.PollRepository, cache *repository.CacheRepository) *PollService {
	return &PollService{repo: repo, cache: cache}
}

func (s *PollService) CreatePoll(ctx context.Context, req *domain.CreatePollRequest) (*domain.Poll, error) {

	if req.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidPoll)
	}
	if len(req.Options) < 2 {
		return nil, fmt.Errorf("%w: at least 2 options are required", ErrInvalidPoll)
	}
	if len(req.Options) > 10 {
		return nil, fmt.Errorf("%w: no more than 10 options are allowed", ErrInvalidPoll)
	}

	poll := &domain.Poll{
		Title:     req.Title,
		ExpiresAt: req.ExpiresAt,
	}

	err := s.repo.CreatePoll(ctx, poll, req.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to create poll: %w", err)
	}
	if err := s.cache.SetPoll(ctx, poll); err != nil {
		slog.Warn("Failed to cache poll", "poll_id", poll.ID, "error", err)
	}

	return poll, nil
}

func (s *PollService) GetPoll(ctx context.Context, id string) (*domain.Poll, error) {

	//get from cache

	cachedPoll, err := s.cache.GetPoll(ctx, id)
	if err != nil {
		slog.Warn("Failed to get from cache", "poll_id", id, "error", err)
	}

	if cachedPoll != nil {
		slog.Info("Cache hit", "poll_id", id)
		return cachedPoll, nil
	}

	slog.Info("Cache miss", "poll_id", id)

	//get from db

	poll, err := s.repo.GetPoll(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get poll by ID: %w", err)
	}
	if poll == nil {
		return nil, ErrPollNotFound
	}

	//save to cache

	if err := s.cache.SetPoll(ctx, poll); err != nil {
		slog.Warn("Failed to cache poll", "poll_id", id, "error", err)
	}

	return poll, nil
}

func (s *PollService) ListPolls(ctx context.Context, page, pageSize int) ([]domain.Poll, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.ListPolls(ctx, pageSize, offset)
}

func (s *PollService) InvalidateCache(ctx context.Context, id string) error {
	return s.cache.DeletePoll(ctx, id)
}
