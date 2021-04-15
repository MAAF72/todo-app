package models

import (
	"todo-app/database"
)

type TodoModel struct {
}

func (m *TodoModel) Create(authorID int64, todo database.Todo) (int64, error) {
	todo.AuthorId = authorID
	_, err := database.DB.Engine.Table("todo").Insert(&todo)

	return todo.Id, err
}

func (m *TodoModel) GetAll(authorID int64) ([]database.Todo, error) {
	var todos []database.Todo

	err := database.DB.Engine.Table("todo").Where("author_id=?", authorID).Find(&todos)

	return todos, err
}

func (m *TodoModel) Get(todoID int64, authorID int64) (*database.Todo, error) {
	todo := &database.Todo{}

	if _, err := database.DB.Engine.Table("todo").ID(todoID).Get(todo); err != nil {
		return nil, err
	}

	if todo.AuthorId != authorID {
		return nil, nil
	}

	return todo, nil
}

func (m *TodoModel) Update(todoID int64, authorID int64, updatedTodo *map[string]interface{}) (bool, error) {
	cols := []string{}

	todo, err := m.Get(todoID, authorID)

	if todo == nil {
		return false, err
	}

	if val, exist := (*updatedTodo)["completed"].(bool); exist {
		cols = append(cols, "completed")
		todo.Completed = val
	}

	if val, exist := (*updatedTodo)["name"].(string); exist {
		cols = append(cols, "name")
		todo.Name = val
	}

	if val, exist := (*updatedTodo)["description"].(string); exist {
		cols = append(cols, "description")
		todo.Description = val
	}

	_, err = database.DB.Engine.Table("todo").ID(todoID).Cols(cols...).UseBool("completed").Update(todo)

	return err == nil, err
}

func (m *TodoModel) Delete(todoID int64, authorID int64) (bool, error) {
	todo, err := m.Get(todoID, authorID)

	if todo == nil {
		return false, err
	}

	_, err = database.DB.Engine.Table("todo").ID(todo.Id).Delete(todo)

	return err == nil, err
}
