package cache

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

type SessionStore struct {
	vdb valkey.Client
}

func (s *SessionStore) GetUser(ctx context.Context, key string) (string, error) {
	return s.vdb.Do(ctx, s.vdb.B().Get().Key(key).Build()).ToString()
}

func (s *SessionStore) Set(ctx context.Context, key, userID string, exp time.Time) error {
	ttl := time.Until(exp).Seconds()
	return s.vdb.Do(ctx, s.vdb.B().Setex().Key(key).Seconds(int64(ttl)).Value(userID).Build()).Error()
}

func (s *SessionStore) Delete(ctx context.Context, key string) error {
	return s.vdb.Do(ctx, s.vdb.B().Del().Key(key).Build()).Error()
}
