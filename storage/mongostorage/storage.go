package mongostorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
	storage2 "twitter/storage"
)

const dbName = "blog_app_db"
const collectionName = "posts"

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
			Keys: bsonx.Doc{
				{Key: "authorId", Value: bsonx.Int32(1)},
				{Key: "_id", Value: bsonx.Int32(-1)},
			},
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
		res, err := s.posts.InsertOne(ctx, data)
		fmt.Println(res)
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
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return storage2.PostData{}, fmt.Errorf("invalid id - %w", storage2.StorageError)
	}
	err = s.posts.FindOne(ctx, bson.M{"_id": objectId}).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.PostData{}, fmt.Errorf("no posts with id %v - %w", id, storage2.ErrorNotFound)
		}
		return storage2.PostData{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
	}
	return result, nil
}

func (s *storage) GetPostsByUserId(ctx context.Context, userId string, pageSize int, pageId string) (storage2.PostsByUser, error) {
	var posts []storage2.PostData
	var post storage2.PostData
	opts := options.Find()
	opts.SetSort(bson.D{
		{"authorId", 1},
		{"_id", -1},
	})
	opts.SetLimit(int64(pageSize))

	cursor, err := s.posts.Find(ctx, bson.M{"authorId": userId, "_id": "$gte: " + pageId}, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.PostsByUser{}, fmt.Errorf("no posts with userId %v and pageId %v - %w", userId, pageId, storage2.ErrorNotFound)
		}
		return storage2.PostsByUser{}, fmt.Errorf("something went wrong - %w", storage2.StorageError)
	}
	for cursor.Next(ctx) {
		err := cursor.Decode(&post)
		if err != nil {
			return storage2.PostsByUser{}, err
		}
		posts = append(posts, post)
	}
	return storage2.PostsByUser{Posts: posts, NextPageId: post.Id.String()}, nil
}
