package main

import (
	"fmt"

	"todo-app/config"
	"todo-app/controllers"
	"todo-app/database"
	"todo-app/services"

	"github.com/kataras/iris/v12"
	"github.com/spf13/viper"
)

var conf config.Configuration

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	confErr := viper.ReadInConfig()
	if confErr != nil {
		fmt.Println("Error loading config file")
	}

	confErr = viper.Unmarshal(&conf)
	if confErr != nil {
		fmt.Println("Error unmarshalling config file")
	}

	database.DB.Init(conf.Database)

	app := iris.New()

	services.JWTServices.Init(conf.JWT)
	controllers.GetTodosParty(app)
	controllers.GetAuthParty(app)

	app.Listen(fmt.Sprintf(":%d", conf.Server.Port))
}
