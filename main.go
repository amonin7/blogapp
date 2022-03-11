package main

import (
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
	"time"
	handler2 "twitter/handler"
	"twitter/storage/mongostorage"
	"twitter/storage/rediscachedstorage"
)

func NewServer() *http.Server {

	mongoUrl := os.Getenv("MONGO_URL")
	mongoStorage := mongostorage.DatabaseStorage(mongoUrl)
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	cachedStorage := rediscachedstorage.NewStorage(mongoStorage, redisClient)
	router := handler2.CreateRouterFromStorage(cachedStorage)

	return &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func main() {
	srv := NewServer()
	log.Printf("Start serving on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
