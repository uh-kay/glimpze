package cache

import (
	"context"
	"time"

	"github.com/uh-kay/glimpze/store"
	"github.com/valkey-io/valkey-go"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
	Sessions interface {
		GetUser(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key, userID string, exp time.Time) error
		Delete(ctx context.Context, key string) error
	}
}

func NewValkeyStorage(vdb valkey.Client) Storage {
	return Storage{
		Users:    &UserStore{vdb: vdb},
		Sessions: &SessionStore{vdb: vdb},
	}
}
