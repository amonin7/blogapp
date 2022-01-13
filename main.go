package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type PublicationRequestData struct {
	Text string `json:"text"`
}

type HttpHandler struct {
	ds DataSource
}

func getRandomKey() string {
	alphaBet := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Shuffle(len(alphaBet), func(i, j int) {
		alphaBet[i], alphaBet[j] = alphaBet[j], alphaBet[i]
	})
	id := string(alphaBet[:10])
	return id
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
	postId := getRandomKey()

	postData := PostData{postId, publicationData.Text, userId, time.Now().String()}
	err = h.ds.save(postData)
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

	post, err := h.ds.getPostById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	posts := h.ds.getPostsByUserId(userId)

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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hello from Server!"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain")
}

func NewServer() *http.Server {
	r := mux.NewRouter()

	ids := &InmemoryDataSource{
		idToPost:      make(map[string]PostData),
		userIdToPosts: make(map[string][]PostData),
	}

	handler := &HttpHandler{
		ds: ids,
	}

	r.HandleFunc("/", handleRoot)
	r.HandleFunc("/api/v1/posts", handler.handlePublication).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:\\w{10}}", handler.handleGetPublication).Methods(http.MethodGet)
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
