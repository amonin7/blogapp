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
	Save(ctx context.Context, data PostData) error
	GetPostById(ctx context.Context, id string) (PostData, error)
	GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (PostsByUser, error)
}
