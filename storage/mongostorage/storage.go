package mongostorage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/description"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readconcern"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	_ "go.mongodb.org/mongo-driver/mongo/writeconcern"
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
		_, err := s.posts.InsertOne(ctx, data)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			return fmt.Errorf("something went wrong - %w", storage2.CommonStorageError)
		}

		return nil
	}
	return fmt.Errorf("too much attempts during inserting - %w", storage2.ErrorCollision)
}

func (s *storage) GetPostById(ctx context.Context, id string) (storage2.PostData, error) {
	var result storage2.PostData
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return storage2.PostData{}, fmt.Errorf("invalid id - %w", storage2.CommonStorageError)
	}
	err = s.posts.FindOne(ctx, bson.M{"_id": objectId}).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.PostData{}, fmt.Errorf("no posts with id %v - %w", id, storage2.ErrorNotFound)
		}
		return storage2.PostData{}, fmt.Errorf("something went wrong - %w", storage2.CommonStorageError)
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
	var cursor *mongo.Cursor
	var err error
	if pageId == "" {
		cursor, err = s.posts.Find(ctx, bson.M{"authorId": userId}, opts)
	} else {
		objectId, err := primitive.ObjectIDFromHex(pageId)
		if err != nil {
			return storage2.PostsByUser{}, fmt.Errorf("invalid id - %w", storage2.CommonStorageError)
		}
		cursor, err = s.posts.Find(ctx, bson.M{"authorId": userId, "_id": bson.M{"$lt": objectId}}, opts)
	}

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage2.PostsByUser{}, fmt.Errorf("no posts with userId %v and pageId %v - %w", userId, pageId, storage2.ErrorNotFound)
		}
		return storage2.PostsByUser{}, fmt.Errorf("something went wrong - %w", storage2.CommonStorageError)
	}
	hasPost := false
	for cursor.Next(ctx) {
		if !hasPost {
			hasPost = true
		}
		err := cursor.Decode(&post)
		if err != nil {
			return storage2.PostsByUser{}, err
		}
		posts = append(posts, post)
	}
	if !hasPost {
		return storage2.PostsByUser{Posts: posts, NextPageId: primitive.NilObjectID}, nil
	} else {
		return storage2.PostsByUser{Posts: posts, NextPageId: post.Id}, nil
	}
}

func (s *storage) Update(ctx context.Context, data storage2.PostData) error {
	update := bson.M{"$set": bson.M{"text": data.Text}}
	_, err := s.posts.UpdateByID(ctx, data.Id, update)
	if err != nil {
		return fmt.Errorf("something went wrong - %w, %v", storage2.CommonStorageError, err)
	}
	return nil
}
