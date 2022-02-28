package rediscachedstorage

import (
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
	"twitter/storage"
)

func NewStorage(persistentStorage storage.Storage, client *redis.Client) *Storage {
	return &Storage{
		client:            client,
		persistentStorage: persistentStorage,
	}
}

type Storage struct {
	client            *redis.Client
	persistentStorage storage.Storage
}
