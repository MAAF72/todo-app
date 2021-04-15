package models

import (
	"todo-app/database"
)

type AuthorModel struct {
}

func (m *AuthorModel) Create(username string, hash string) (int64, error) {
	author := database.Author{}
	author.Username = username
	author.Hash = hash

	_, err := database.DB.Engine.Table("author").Insert(&author)

	return author.Id, err
}

func (m *AuthorModel) Get(username string) (*database.Author, error) {
	author := &database.Author{Username: username}

	has, err := database.DB.Engine.Table("author").Get(author)

	if err != nil || !has {
		return nil, err
	}

	return author, nil
}
