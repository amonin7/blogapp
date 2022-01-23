package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"twitter/storage"
	"twitter/storage/mongostorage"
)

type PublicationRequestData struct {
	Text string `json:"text"`
}

type HttpHandler struct {
	ds storage.DataSource
}

func isValidateUserId(userId string) bool {
	re := regexp.MustCompile(`[0-9a-z]+`)
	matches := re.FindAllString(userId, -1)
	if userId == "" {
		return false
	} else if len(matches) != 1 {
		return false
	} else {
		return true
	}
}

func (h *HttpHandler) handlePublication(w http.ResponseWriter, r *http.Request) {
	var publicationData PublicationRequestData
	err := json.NewDecoder(r.Body).Decode(&publicationData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId := r.Header.Get("System-Design-User-Id")
	if !isValidateUserId(userId) {
		http.Error(w, "Provided userId is not valid", http.StatusUnauthorized)
		return
	}
	postData := storage.PostData{
		Id:        primitive.NewObjectID(),
		Text:      publicationData.Text,
		AuthorId:  userId,
		CreatedAt: time.Now().String(),
	}
	err = h.ds.Save(r.Context(), postData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, err := json.Marshal(postData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(rawResponse)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (h *HttpHandler) handleGetPublication(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	postId := parts[len(parts)-1]

	post, err := h.ds.GetPostById(r.Context(), postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	rawResponse, err := json.Marshal(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(rawResponse)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (h *HttpHandler) handleGetPublicationsByUser(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		http.Error(w, "No userId in url", http.StatusBadRequest)
		return
	}
	userId := parts[len(parts)-2]

	pageSizeParam := r.URL.Query()["size"]
	pageSize := 10
	pageIdParam := r.URL.Query()["page"]
	pageId := ""
	if len(pageSizeParam) > 1 {
		http.Error(w, "More than 1 query param \"size\"", http.StatusBadRequest)
		return
	} else if len(pageSizeParam) == 1 {
		i, err := strconv.Atoi(pageSizeParam[0])
		if err != nil {
			http.Error(w, "query param \"size\" should be integer", http.StatusBadRequest)
			return
		}
		pageSize = i
	}
	if len(pageIdParam) > 1 {
		http.Error(w, "More than 1 query param \"page\"", http.StatusBadRequest)
		return
	} else if len(pageIdParam) == 1 {
		pageId = pageIdParam[0]
	}

	posts, err := h.ds.GetPostsByUserId(r.Context(), userId, pageSize, pageId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(rawResponse)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (h *HttpHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hello from Server!"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain")
}

func (h *HttpHandler) handlePing(w http.ResponseWriter, r *http.Request) {
	return
}

func NewServer() *http.Server {
	r := mux.NewRouter()

	mongoUrl := os.Getenv("MONGO_URL")
	mongoStorage := mongostorage.DatabaseStorage(mongoUrl)

	handler := &HttpHandler{
		ds: mongoStorage,
	}

	r.HandleFunc("/", handler.handleRoot)
	r.HandleFunc("/maintenance/ping", handler.handlePing).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts", handler.handlePublication).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:\\w+}", handler.handleGetPublication).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{userId:\\w+}/posts", handler.handleGetPublicationsByUser).Methods(http.MethodGet)

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
