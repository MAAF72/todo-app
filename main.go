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
	var db_err error
	engine, db_err = xorm.NewEngine(
		DB_DRIVER,
		fmt.Sprintf(
			"user=%s dbname=%s password=%s port=%d sslmode=disable",
			DB_USER, DB_NAME, DB_PASSWORD, DB_PORT,
		),
	)

	if db_err != nil {
		fmt.Println("Error when connecting to database")
		fmt.Println(db_err)
		return
	}

	db_err = engine.Sync2(new(Todo), new(Author))

	if db_err != nil {
		fmt.Println("Error when syning with database")
		fmt.Println(db_err)
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

		todosAPI.Get("/", list_todo)
		todosAPI.Post("/", create_todo)
		todosAPI.Patch("/{id:int}", update_todo)
		todosAPI.Delete("/{id:int}", delete_todo)
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

	read_err := ctx.ReadJSON(&data)

	if read_err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	hashed_password, hash_err := bcrypt.GenerateFromPassword([]byte(data["Password"]), bcrypt.DefaultCost)

	if hash_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	author := Author{}
	author.Username = data["Username"]
	author.Hash = string(hashed_password)

	_, query_err := engine.Table("author").Insert(&author)

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.StatusCode(iris.StatusCreated)
}

func login(ctx iris.Context) {
	var data map[string]string

	read_err := ctx.ReadJSON(&data)

	if read_err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	author := Author{Username: data["Username"]}
	has, query_err := engine.Table("author").Get(&author)

	if query_err != nil {
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

	if bcrypt.CompareHashAndPassword([]byte(author.Hash), []byte(data["Password"])) != nil {
		ctx.JSON(iris.Map{
			"success": false,
			"message": "Wrong password",
		})
		ctx.StatusCode(iris.StatusOK)
		return
	}

	claims := AuthClaims{Id: author.Id}

	token, sign_err := signer.Sign(claims)

	if sign_err != nil {
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
	logout_err := ctx.Logout()

	if logout_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.StatusCode(iris.StatusOK)
}

func create_todo(ctx iris.Context) {
	var todo Todo
	claims := jwt.Get(ctx).(*AuthClaims)

	read_err := ctx.ReadJSON(&todo)

	if read_err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	todo.AuthorId = claims.Id

	_, query_err := engine.Table("todo").Insert(&todo)

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.StatusCode(iris.StatusCreated)
}

func list_todo(ctx iris.Context) {
	var todos []Todo

	claims := jwt.Get(ctx).(*AuthClaims)

	_, query_err := engine.Table("todo").Where("author_id=?", claims.Id).Get(&todos)

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.JSON(todos)

	ctx.StatusCode(iris.StatusOK)
}

func update_todo(ctx iris.Context) {
	var todo Todo

	todo_id, query_err := ctx.Params().GetInt64("id")

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	todo.Id = todo_id

	has, query_err := engine.Table("todo").ID(todo.Id).Get(&todo)

	if query_err != nil {
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

	read_err := ctx.ReadJSON(&todo)

	if read_err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	_, query_err = engine.Table("todo").ID(todo.Id).Cols("name", "description", "completed").UseBool("completed").Update(&todo)

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.StatusCode(iris.StatusOK)
}

func delete_todo(ctx iris.Context) {
	var todo Todo

	todo_id, params_err := ctx.Params().GetInt64("id")

	if params_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	todo.Id = todo_id

	has, query_err := engine.Table("todo").ID(todo.Id).Get(&todo)

	if query_err != nil {
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

	_, query_err = engine.Table("todo").ID(todo.Id).Delete(&todo)

	if query_err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}

	ctx.StatusCode(iris.StatusOK)
}
