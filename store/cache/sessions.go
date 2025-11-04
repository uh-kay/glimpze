package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uh-kay/glimpze/store"
	"github.com/valkey-io/valkey-go"
)

type SessionStore struct {
	vdb valkey.Client
}

func (s *SessionStore) GetUser(ctx context.Context, key string) (*store.User, error) {
	data, err := s.vdb.Do(ctx, s.vdb.B().Get().Key(key).Build()).ToString()
	if valkey.IsValkeyNil(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *SessionStore) Set(ctx context.Context, key, userID string, exp time.Time) error {
	ttl := time.Until(exp).Seconds()
	return s.vdb.Do(ctx, s.vdb.B().Setex().Key(key).Seconds(int64(ttl)).Value(userID).Build()).Error()
}

func (s *SessionStore) Delete(ctx context.Context, key string) error {
	return s.vdb.Do(ctx, s.vdb.B().Del().Key(key).Build()).Error()
}
