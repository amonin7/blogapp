package rediscachedstorage

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
	"twitter/storage"
)

const cacheTTL = 10 * time.Second

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
	fullKey := s.fullKey(data.Id.Hex())
	rawResponse, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp := s.client.Set(ctx, fullKey, rawResponse, cacheTTL)
	if err := resp.Err(); err != nil {
		log.Printf("Failed to save key %s to redis", fullKey)
		return err
	}
	return nil
}

func (s *Storage) GetPostById(ctx context.Context, id string) (storage.PostData, error) {
	fullKey := s.fullKey(id)
	rawData, err := s.client.Get(ctx, fullKey).Result()
	result := storage.PostData{}
	switch {
	case err == redis.Nil:
	// go to persistence
	case err != nil:
		return result, err
	default:
		log.Println("Successfully loaded key from cache")
		err = json.Unmarshal([]byte(rawData), &result)
		if err != nil {
			return storage.PostData{}, err
		}
		return result, nil
	}

	result, err = s.persistentStorage.GetPostById(ctx, id)
	if err != nil {
		return storage.PostData{}, err
	}

	rawPost, err := json.Marshal(result)
	if err != nil {
		return storage.PostData{}, err
	}

	resp := s.client.Set(ctx, fullKey, string(rawPost), cacheTTL)
	if err := resp.Err(); err != nil {
		log.Printf("Failed to save key %s to redis", fullKey)
		return storage.PostData{}, err
	}

	log.Println("Successfully loaded key from persistence")
	return result, nil
}

func (s *Storage) GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (storage.PostsByUser, error) {
	fullKey := s.fullPostsByUserIdKey(userId, pageSize, pageId)
	rawData, err := s.client.Get(ctx, fullKey).Result()
	result := storage.PostsByUser{}
	switch {
	case err == redis.Nil:
	// go to persistence
	case err != nil:
		// this is similar to returning null
		return result, err
	default:
		log.Println("Successfully loaded key from cache")
		err = json.Unmarshal([]byte(rawData), &result)
		if err != nil {
			return storage.PostsByUser{}, err
		}
		return result, nil
	}

	result, err = s.persistentStorage.GetPostsByUserId(ctx, userId, pageSize, pageId)
	if err != nil {
		return storage.PostsByUser{}, err
	}

	rawPostsByUser, err := json.Marshal(result)
	if err != nil {
		return storage.PostsByUser{}, err
	}

	resp := s.client.Set(ctx, fullKey, string(rawPostsByUser), cacheTTL)
	if err := resp.Err(); err != nil {
		log.Printf("Failed to save key %s to redis", fullKey)
		return storage.PostsByUser{}, err
	}

	log.Println("Successfully loaded key from persistence")
	return result, nil
}

func (s *Storage) Update(ctx context.Context, data storage.PostData) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) fullKey(id string) string {
	return "pd:" + id
}

func (s *Storage) fullPostsByUserIdKey(userId string, pageSize int, pageId string) string {
	return "pd:" + userId + ";" + strconv.Itoa(pageSize) + ";" + pageId
}

var _ storage.Storage = (*Storage)(nil)
