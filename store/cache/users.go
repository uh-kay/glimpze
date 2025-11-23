package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"newsdrop.org/store"
	"github.com/valkey-io/valkey-go"
)

type UserStore struct {
	vdb valkey.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)
	data, err := s.vdb.Do(ctx, s.vdb.B().Get().Key(cacheKey).Build()).ToString()
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

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.vdb.Do(ctx, s.vdb.B().Setex().Key(cacheKey).Seconds(int64(UserExpTime.Seconds())).Value(string(json)).Build()).Error()
}
