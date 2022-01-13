package main

import (
	"errors"
	"sync"
)

type PostData struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AuthorId  string `json:"authorId"`
	CreatedAt string `json:"createdAt"`
}

type DataSource interface {
	save(data PostData) error
	getPostById(id string) (PostData, error)
	getPostsByUserId(userId string) []PostData
}

type InmemoryDataSource struct {
	storageMu     sync.RWMutex
	idToPost      map[string]PostData
	userIdToPosts map[string][]PostData
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

func (ids *InmemoryDataSource) getPostsByUserId(userId string) []PostData {
	val, ok := ids.userIdToPosts[userId]
	if ok {
		return val
	} else {
		return nil
	}
}
