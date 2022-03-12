package handler

import (
	"github.com/gorilla/mux"
	"net/http"
	"twitter/storage"
)

func CreateRouterFromStorage(cachedStorage storage.Storage) *mux.Router {
	handler := &HttpHandler{
		Storage: cachedStorage,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HandleRoot)
	r.HandleFunc("/maintenance/ping", handler.HandlePing).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts", handler.HandlePublication).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:\\w+}", handler.HandleGetPublication).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts/{postId:\\w+}", handler.HandleUpdatePublication).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/users/{userId:\\w+}/posts", handler.HandleGetPublicationsByUser).Methods(http.MethodGet)

	return r
}
