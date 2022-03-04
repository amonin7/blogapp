package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
	handler2 "twitter/handler"
	"twitter/storage/mongostorage"
	"twitter/storage/rediscachedstorage"
)

func NewServer() *http.Server {
	r := mux.NewRouter()

	mongoUrl := os.Getenv("MONGO_URL")
	mongoStorage := mongostorage.DatabaseStorage(mongoUrl)
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	cachedStorage := rediscachedstorage.NewStorage(mongoStorage, redisClient)

	handler := &handler2.HttpHandler{
		Storage: cachedStorage,
	}

	r.HandleFunc("/", handler.HandleRoot)
	r.HandleFunc("/maintenance/ping", handler.HandlePing).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts", handler.HandlePublication).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:\\w+}", handler.HandleGetPublication).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts/{postId:\\w+}", handler.HandleUpdatePublication).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/users/{userId:\\w+}/posts", handler.HandleGetPublicationsByUser).Methods(http.MethodGet)

	return &http.Server{
		Handler:      r,
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
