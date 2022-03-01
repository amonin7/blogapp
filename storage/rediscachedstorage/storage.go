package rediscachedstorage

import (
	"context"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
	"log"
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

func (s *Storage) Save(ctx context.Context, data storage.PostData) error {
	err := s.persistentStorage.Save(ctx, data)
	if err != nil {
		return err
	}
	fullKey := s.fullKey(key)
	resp := s.client.Set(ctx, fullKey, string(url), cacheTTL)
	if err := resp.Err(); err != nil {
		log.Printf("Failed to save key %s to redis", fullKey)
		return "", err
	}

	return key, nil
}

func (s *Storage) GetPostById(ctx context.Context, id string) (storage.PostData, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (storage.PostsByUser, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) Update(ctx context.Context, data storage.PostData) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) fullKey(data storage.PostData) string {
	return "su:" + string(data.Id)
}

var _ storage.Storage = (*Storage)(nil)
