package handler

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"twitter/storage"
)

type PublicationRequestData struct {
	Text string `json:"text"`
}

type HttpHandler struct {
	Storage storage.Storage
}

func isValidUserId(userId string) bool {
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

func (h *HttpHandler) HandlePublication(w http.ResponseWriter, r *http.Request) {
	var publicationData PublicationRequestData
	err := json.NewDecoder(r.Body).Decode(&publicationData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId := r.Header.Get("System-Design-User-Id")
	if !isValidUserId(userId) {
		http.Error(w, "Provided userId is not valid", http.StatusUnauthorized)
		return
	}
	postData := storage.PostData{
		Id:        primitive.NewObjectID(),
		Text:      publicationData.Text,
		AuthorId:  userId,
		CreatedAt: time.Now().String(),
	}
	err = h.Storage.Save(r.Context(), postData)
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

func (h *HttpHandler) HandleGetPublication(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	postId := parts[len(parts)-1]

	post, err := h.Storage.GetPostById(r.Context(), postId)
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

func (h *HttpHandler) HandleGetPublicationsByUser(w http.ResponseWriter, r *http.Request) {
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

	posts, err := h.Storage.GetPostsByUserId(r.Context(), userId, pageSize, pageId)
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

func (h *HttpHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hello from Server!"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain")
}

func (h *HttpHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *HttpHandler) HandleUpdatePublication(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	postId := parts[len(parts)-1]

	post, err := h.Storage.GetPostById(r.Context(), postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var publicationData PublicationRequestData
	err = json.NewDecoder(r.Body).Decode(&publicationData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId := r.Header.Get("System-Design-User-Id")
	if !isValidUserId(userId) {
		http.Error(w, "Provided userId is not valid", http.StatusUnauthorized)
		return
	}

	if userId != post.AuthorId {
		http.Error(w, "This post is published by another user", http.StatusForbidden)
		return
	}

	post.Text = publicationData.Text
	err = h.Storage.Update(r.Context(), post)
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
