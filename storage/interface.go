package storage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	StorageError   = errors.New("storage")
	ErrorCollision = fmt.Errorf("%w.collision", StorageError)
	ErrorNotFound  = fmt.Errorf("%w.not_found", StorageError)
)

type PostData struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Text      string             `json:"text" bson:"text"`
	AuthorId  string             `json:"authorId" bson:"authorId"`
	CreatedAt string             `json:"createdAt" bson:"createdAt"`
}

type PostsByUser struct {
	Posts      []PostData         `json:"posts" bson:"posts"`
	NextPageId primitive.ObjectID `json:"nextPage" bson:"nextPage"`
}

type DataSource interface {
	Save(ctx context.Context, data PostData) error
	GetPostById(ctx context.Context, id string) (PostData, error)
	GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (PostsByUser, error)
	Update(ctx context.Context, data PostData) error
}
