package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	redis "github.com/redis/go-redis/v9"
)

type Session struct {
	UserProp       string `json:"user"`
	ValidUntilProp int64  `json:"validUntil"`
}

// User implements models.Session
func (s *Session) User() string {
	return s.UserProp
}

// ValidUntil implements models.Session
func (s *Session) ValidUntil() time.Time {
	return time.UnixMicro(s.ValidUntilProp)
}

type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(redisClient *redis.Client) *SessionRepository {
	return &SessionRepository{redisClient}
}

// Get implements repository.SessionRepository
func (r *SessionRepository) Get(ctx context.Context, token string) (models.Session, error) {
	res := r.client.Get(ctx, sessionKey(token))
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error reading from redis: %w", err)
	}

	raw, _ := res.Bytes()
	var session Session
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	return &session, nil
}

// Put implements repository.SessionRepository
func (r *SessionRepository) Put(ctx context.Context, token string, userID string, TTL time.Duration) error {
	serialized, err := json.Marshal(&Session{
		UserProp: userID,
		ValidUntilProp: time.Now().Add(TTL).UnixMicro(),
	})
	if err != nil {
		return fmt.Errorf("error serializing session: %w", err)
	}

	if res := r.client.Set(ctx, sessionKey(token), serialized, time.Second * time.Duration(TTL.Seconds())); err != nil {
		return fmt.Errorf("error writing token to redis: %w", res.Err())
	}

	return nil
}

// Remove implements repository.SessionRepository
func (r *SessionRepository) Remove(ctx context.Context, token string) error {
	if err := r.client.Del(ctx, sessionKey(token)).Err(); err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}

	return nil
}

func sessionKey(token string) string {
	return fmt.Sprintf("session::%s", token)
}

var _ models.Session = (*Session)(nil)
var _ repository.SessionRepository = (*SessionRepository)(nil)
