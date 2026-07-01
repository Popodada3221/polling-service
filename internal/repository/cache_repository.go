package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"polling-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheRepository(client *redis.Client, ttlSeconds int) *CacheRepository {
	return &CacheRepository{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

func (r *CacheRepository) GetPoll(ctx context.Context, id string) (*domain.Poll, error) {
	key := fmt.Sprintf("poll:%s", id)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var poll domain.Poll
	if err := json.Unmarshal(data, &poll); err != nil {
		return nil, fmt.Errorf("failed to unmarshall poll: %w", err)
	}

	return &poll, nil
}

func (r *CacheRepository) SetPoll(ctx context.Context, poll *domain.Poll) error {
	key := fmt.Sprintf("poll:%s", poll.ID)

	data, err := json.Marshal(poll)
	if err != nil {
		return fmt.Errorf("failed to marshall poll: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (r *CacheRepository) DeletePoll(ctx context.Context, id string) error {
	key := fmt.Sprintf("poll:%s", id)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

func (r *CacheRepository) DeletePollsPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
