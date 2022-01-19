package storage

import (
	"context"
	"errors"
	"fmt"
)

var (
	StorageError   = errors.New("storage")
	ErrorCollision = fmt.Errorf("%w.collision", StorageError)
	ErrorNotFound  = fmt.Errorf("%w.not_found", StorageError)
)

type PostData struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type PostsByUser struct {
	Posts      []PostData `json:"posts"`
	NextPageId string     `json:"nextPage"`
}

type DataSource interface {
	save(ctx context.Context, data PostData) error
	getPostById(ctx context.Context, id string) (PostData, error)
	getPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (PostsByUser, error)
}
