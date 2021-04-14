package main

import (
	"fmt"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"

	_ "github.com/lib/pq"
)

var engine *xorm.Engine
var signer *jwt.Signer
var secret = []byte("INI_SECRET_KEY")
var signature = jwt.HS256

const (
	DB_DRIVER   = "postgres"
	DB_USER     = "postgres"
	DB_NAME     = "postgres"
	DB_PASSWORD = "123456"
	DB_PORT     = 5432
)

func main() {
	var dbErr error
	engine, dbErr = xorm.NewEngine(
		DB_DRIVER,
		fmt.Sprintf(
			"user=%s dbname=%s password=%s port=%d sslmode=disable",
			DB_USER, DB_NAME, DB_PASSWORD, DB_PORT,
		),
	)

	if dbErr != nil {
		fmt.Println("Error when connecting to database")
		fmt.Println(dbErr)
		return
	}

	dbErr = engine.Sync2(new(Todo), new(Author))

	if dbErr != nil {
		fmt.Println("Error when syning with database")
		fmt.Println(dbErr)
		return
	}

	app := iris.New()

	signer = jwt.NewSigner(signature, secret, 10*time.Minute)
	verifier := jwt.NewVerifier(signature, secret).WithDefaultBlocklist()

	verifyMiddleware := verifier.Verify(func() interface{} { return new(AuthClaims) })

	todosAPI := app.Party("/todos")
	{
		todosAPI.Use(iris.Compression)
		todosAPI.Use(verifyMiddleware)

		todosAPI.Get("/", listTodo)
		todosAPI.Post("/", createTodo)
		todosAPI.Patch("/{id:int}", updateTodo)
		todosAPI.Delete("/{id:int}", deleteTodo)
	}

	authAPI := app.Party("/auth")
	{
		authAPI.Use(iris.Compression)
		authAPI.Post("/register", register)
		authAPI.Post("/login", login)
		authAPI.Get("/logout", verifyMiddleware, logout)
	}

	app.Listen(":8080")
}

type AuthClaims struct {
	Id int64
}

type Author struct {
	Id       int64
	Username string
	Hash     string
	Created  time.Time `xorm:"created"`
}

type Todo struct {
	Id          int64
	Name        string
	Description string `xorm:"text"`
	Completed   bool
	AuthorId    int64
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
}

func register(ctx iris.Context) {
	var data map[string]string

	readErr := ctx.ReadJSON(&data)

	if readErr != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	hashedPass, hashErr := bcrypt.GenerateFromPassword([]byte(data["password"]), bcrypt.DefaultCost)

	if hashErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	author := Author{}
	author.Username = data["username"]
	author.Hash = string(hashedPass)

	_, queryErr := engine.Table("author").Insert(&author)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusCreated)
}

func login(ctx iris.Context) {
	var data map[string]string

	readErr := ctx.ReadJSON(&data)

	if readErr != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	author := Author{Username: data["username"]}
	has, queryErr := engine.Table("author").Get(&author)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	if !has {
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

	claims := AuthClaims{Id: author.Id}

	token, signErr := signer.Sign(claims)

	if signErr != nil {
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
	logoutErr := ctx.Logout()

	if logoutErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusOK)
}

func createTodo(ctx iris.Context) {
	var todo Todo
	claims := jwt.Get(ctx).(*AuthClaims)

	readErr := ctx.ReadJSON(&todo)

	if readErr != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	todo.AuthorId = claims.Id

	_, queryErr := engine.Table("todo").Insert(&todo)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusCreated)
}

func listTodo(ctx iris.Context) {
	var todos []Todo

	claims := jwt.Get(ctx).(*AuthClaims)

	queryErr := engine.Table("todo").Where("author_id=?", claims.Id).Find(&todos)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(todos)

	ctx.StatusCode(iris.StatusOK)
}

func updateTodo(ctx iris.Context) {
	var todo Todo

	todoID, queryErr := ctx.Params().GetInt64("id")

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	todo.Id = todoID

	has, queryErr := engine.Table("todo").ID(todo.Id).Get(&todo)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	if !has {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Todo activity not found",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	claims := jwt.Get(ctx).(*AuthClaims)

	if todo.AuthorId != claims.Id {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Please update your own todo activity",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	readErr := ctx.ReadJSON(&todo)

	if readErr != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	_, queryErr = engine.Table("todo").ID(todo.Id).Cols("name", "description", "completed").UseBool("completed").Update(&todo)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusOK)
}

func deleteTodo(ctx iris.Context) {
	var todo Todo

	todoID, paramsErr := ctx.Params().GetInt64("id")

	if paramsErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	todo.Id = todoID

	has, queryErr := engine.Table("todo").ID(todo.Id).Get(&todo)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	if !has {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Todo activity not found",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	claims := jwt.Get(ctx).(*AuthClaims)

	if todo.AuthorId != claims.Id {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Please delete your own todo activity",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	_, queryErr = engine.Table("todo").ID(todo.Id).Delete(&todo)

	if queryErr != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
	})

	ctx.StatusCode(iris.StatusOK)
}
