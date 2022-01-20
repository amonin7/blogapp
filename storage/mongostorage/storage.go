package mongostorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
	storage2 "twitter/storage"
)

const dbName = "blog_app_db"
const collectionName = "test"

type storage struct {
	posts *mongo.Collection
}

func DatabaseStorage(mongoUrl string) *storage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		panic(err)
	}

	collection := client.Database(dbName).Collection(collectionName)
	ensureIndexes(ctx, collection)

	return &storage{
		posts: collection,
	}
}

// TODO: Add correct  indexes
func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "_id", Value: bsonx.Int32(1)}},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes - %w", err))
	}
}

func (s *storage) Save(ctx context.Context, data storage2.PostData) error {
	for attempt := 0; attempt < 5; attempt++ {
		_, err := s.posts.InsertOne(ctx, data)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return fmt.Errorf("something went wrong - %w", storage2.StorageError)
		}

		return nil
	}
	return fmt.Errorf("too much attempts during inserting - %w", storage2.ErrorCollision)
}

func (s *storage) GetPostById(ctx context.Context, id string) (storage2.PostData, error) {
	var result storage2.PostData
	err := s.posts.FindOne(ctx, bson.M{"id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.PostData{}, fmt.Errorf("no posts with id %v - %w", id, storage2.ErrorNotFound)
		}
		return storage2.PostData{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
	}
	return result, nil
}

func (s *storage) GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (storage2.PostsByUser, error) {
	return storage2.PostsByUser{}, nil
}
