package controllers

import (
	"todo-app/models"
	"todo-app/services"

	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
)

var authorModel models.AuthorModel

func GetAuthParty(app *iris.Application) {
	authAPI := app.Party("/auth")
	{
		authAPI.Use(iris.Compression)
		authAPI.Post("/register", register)
		authAPI.Post("/login", login)
		authAPI.Get("/logout", services.JWTServices.VerifyMiddleware, logout)
	}
}

func register(ctx iris.Context) {
	var data map[string]string

	err := ctx.ReadJSON(&data)

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcrypt.DefaultCost)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	id, err := authorModel.Create(data["username"], string(hashedPass))

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": id != 0,
	})

	ctx.StatusCode(iris.StatusCreated)
}

func login(ctx iris.Context) {
	var data map[string]string

	err := ctx.ReadJSON(&data)

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	author, err := authorModel.Get(data["username"])

	if author == nil {
		ctx.JSON(iris.Map{
			"succes":  false,
			"message": "Username not found",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(author.Hash), []byte(data["password"])) != nil {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Wrong password",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	claims := services.AuthClaims{Id: author.Id}

	token, err := services.JWTServices.Signer.Sign(claims)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
		"token":   string(token),
	})

	ctx.StatusCode(iris.StatusOK)
}

func logout(ctx iris.Context) {
	err := ctx.Logout()

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusOK)
}
