package main

import (
	"errors"
	"math/rand"
	"sync"
)

type PostData struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type PostsByUser struct {
	Posts      []PostData `json:"posts"`
	NextPageId string     `json:"nextPage"`
}

type DataSource interface {
	save(data PostData) error
	getPostById(id string) (PostData, error)
	getPostsByUserId(userId string, pageSize int, pageId string) (PostsByUser, error)
}

type InmemoryDataSource struct {
	storageMu        sync.RWMutex
	idToPost         map[string]PostData
	userIdToPosts    map[string][]PostData
	pageIdToOffset   map[string]int
	pageIdToPageSize map[string]int
}

func (ids *InmemoryDataSource) save(data PostData) error {
	_, ok := ids.idToPost[data.Id]
	if ok {
		return errors.New("this id is already in storage")
	} else {
		ids.idToPost[data.Id] = data
		val, _ := ids.userIdToPosts[data.AuthorId]
		val = append(val, data)
		ids.userIdToPosts[data.AuthorId] = val
		return nil
	}
}

func (ids *InmemoryDataSource) getPostById(id string) (PostData, error) {
	val, ok := ids.idToPost[id]
	if ok {
		return val, nil
	} else {
		return PostData{"", "", "", ""}, errors.New("There is no post with id " + id)
	}
}

func (ids *InmemoryDataSource) getPostsByUserId(userId string, pageSize int, pageId string) (PostsByUser, error) {
	val, ok := ids.userIdToPosts[userId]
	if pageId == "" {
		if ok {
			if len(val) <= pageSize {
				return PostsByUser{val, ""}, nil
			} else {
				newPageId := getRandomKey()
				ids.pageIdToPageSize[newPageId] = pageSize
				ids.pageIdToOffset[newPageId] = pageSize
				return PostsByUser{val[:pageSize], newPageId}, nil
			}
		} else {
			return PostsByUser{[]PostData{}, ""}, nil
		}
	} else {
		if ok {
			oldSize, ok2 := ids.pageIdToPageSize[pageId]
			if ok2 {
				if oldSize != pageSize {
					return PostsByUser{[]PostData{}, ""}, errors.New("provided pageid is not equal to current pageid")
				}
				val = val[ids.pageIdToOffset[pageId]:]
				if len(val) <= pageSize {
					return PostsByUser{val, ""}, nil
				} else {
					newPageId := getRandomKey()
					ids.pageIdToPageSize[newPageId] = pageSize
					ids.pageIdToOffset[newPageId] = ids.pageIdToOffset[pageId] + pageSize
					return PostsByUser{val[:pageSize], newPageId}, nil
				}
			} else {
				return PostsByUser{[]PostData{}, ""}, errors.New("page was not found")
			}
		} else {
			return PostsByUser{[]PostData{}, ""}, errors.New("this user has no posts yet")
		}
	}
}

func getRandomKey() string {
	alphaBet := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Shuffle(len(alphaBet), func(i, j int) {
		alphaBet[i], alphaBet[j] = alphaBet[j], alphaBet[i]
	})
	id := string(alphaBet[:10])
	return id
}
