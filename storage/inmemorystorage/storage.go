package inmemorystorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
	"twitter/generator"
	"twitter/storage"
)

type InmemoryDataSource struct {
	StorageMu        sync.RWMutex
	IdToPost         map[string]storage.PostData
	UserIdToPosts    map[string][]storage.PostData
	PageIdToOffset   map[string]int
	PageIdToPageSize map[string]int
}

func (ids *InmemoryDataSource) Save(ctx context.Context, data storage.PostData) error {
	for attempt := 0; attempt < 5; attempt++ {
		_, ok := ids.IdToPost[data.Id.String()]
		if ok {
			continue
		} else {
			ids.IdToPost[data.Id.String()] = data
			val, _ := ids.UserIdToPosts[data.AuthorId]
			val = append(val, data)
			ids.UserIdToPosts[data.AuthorId] = val
			return nil
		}
	}
	return fmt.Errorf("too much attempts during inserting - %w", storage.ErrorCollision)
}

func (ids *InmemoryDataSource) GetPostById(ctx context.Context, id string) (storage.PostData, error) {
	val, ok := ids.IdToPost[id]
	if ok {
		return val, nil
	} else {
		return storage.PostData{}, fmt.Errorf("no posts with id %v - %w", id, storage.ErrorNotFound)
	}
}

func (ids *InmemoryDataSource) GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (storage.PostsByUser, error) {
	val, ok := ids.UserIdToPosts[userId]
	if pageId == "" {
		if ok {
			if len(val) <= pageSize {
				return storage.PostsByUser{Posts: val}, nil
			} else {
				newPageId := generator.GetRandomKey()
				ids.PageIdToPageSize[newPageId] = pageSize
				ids.PageIdToOffset[newPageId] = pageSize
				objectId, err := primitive.ObjectIDFromHex(newPageId)
				if err != nil {
					return storage.PostsByUser{}, fmt.Errorf("invalid id - %w", storage.StorageError)
				}
				return storage.PostsByUser{Posts: val[:pageSize], NextPageId: objectId}, nil
			}
		} else {
			return storage.PostsByUser{Posts: []storage.PostData{}}, nil
		}
	} else {
		if ok {
			oldSize, ok2 := ids.PageIdToPageSize[pageId]
			if ok2 {
				if oldSize != pageSize {
					return storage.PostsByUser{Posts: []storage.PostData{}}, errors.New("provided pageid is not equal to current pageid")
				}
				val = val[ids.PageIdToOffset[pageId]:]
				if len(val) <= pageSize {
					return storage.PostsByUser{Posts: val}, nil
				} else {
					newPageId := generator.GetRandomKey()
					ids.PageIdToPageSize[newPageId] = pageSize
					ids.PageIdToOffset[newPageId] = ids.PageIdToOffset[pageId] + pageSize
					return storage.PostsByUser{Posts: val[:pageSize]}, nil
				}
			} else {
				return storage.PostsByUser{Posts: []storage.PostData{}}, errors.New("page was not found")
			}
		} else {
			return storage.PostsByUser{Posts: []storage.PostData{}}, errors.New("this user has no posts yet")
		}
	}
}
