package controllers

import (
	"todo-app/database"
	"todo-app/models"
	"todo-app/services"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

var todoModel models.TodoModel

func GetTodosParty(app *iris.Application) {
	todosAPI := app.Party("/todos")
	{
		todosAPI.Use(iris.Compression)
		todosAPI.Use(services.JWTServices.VerifyMiddleware)

		todosAPI.Get("/", listTodo)
		todosAPI.Post("/", createTodo)
		todosAPI.Patch("/{id:int}", updateTodo)
		todosAPI.Delete("/{id:int}", deleteTodo)
	}
}

func createTodo(ctx iris.Context) {
	var todo database.Todo

	claims := jwt.Get(ctx).(*services.AuthClaims)

	err := ctx.ReadJSON(&todo)

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	id, err := todoModel.Create(claims.Id, todo)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": id != 0,
	})

	ctx.StatusCode(iris.StatusCreated)
}

func listTodo(ctx iris.Context) {
	claims := jwt.Get(ctx).(*services.AuthClaims)

	res, err := todoModel.GetAll(claims.Id)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	if res != nil {
		ctx.JSON(res)
	} else {
		ctx.JSON([]string{})
	}

	ctx.StatusCode(iris.StatusOK)
}

func updateTodo(ctx iris.Context) {
	var todo map[string]interface{}
	//var todo database.Todo
	todoID, err := ctx.Params().GetInt64("id")

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	claims := jwt.Get(ctx).(*services.AuthClaims)

	err = ctx.ReadJSON(&todo)

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	ok, err := todoModel.Update(todoID, claims.Id, &todo)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": ok,
	})

	ctx.StatusCode(iris.StatusOK)
}

func deleteTodo(ctx iris.Context) {
	todoID, err := ctx.Params().GetInt64("id")

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	claims := jwt.Get(ctx).(*services.AuthClaims)

	ok, err := todoModel.Delete(todoID, claims.Id)

	ctx.JSON(iris.Map{
		"success": ok,
	})

	ctx.StatusCode(iris.StatusOK)
}
