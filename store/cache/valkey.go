package cache

import (
	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(addr, pw string, db int) (valkey.Client, error) {
	return valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{"localhost:6379"},
		Password:    pw,
		SelectDB:    db,
	})
}
