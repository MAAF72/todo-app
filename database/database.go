package database

import (
	"fmt"
	"time"
	"todo-app/config"

	"xorm.io/xorm"

	_ "github.com/lib/pq"
)

var DB Database

type Database struct {
	Engine *xorm.Engine
	Conf   config.DatabaseConfiguration
}

func (d *Database) Init(conf config.DatabaseConfiguration) {
	d.Conf = conf
	d.CreateEngine()
	d.SyncDatabase()
}

func (d *Database) CreateEngine() {
	var err error
	d.Engine, err = xorm.NewEngine(
		d.Conf.Driver,
		fmt.Sprintf(
			"user=%s dbname=%s password=%s port=%d sslmode=%s",
			d.Conf.User, d.Conf.Name, d.Conf.Password, d.Conf.Port, d.Conf.SSL,
		),
	)

	if err != nil {
		fmt.Println("Error when connecting to database")
		fmt.Println(err)
	}
}

func (d *Database) SyncDatabase() {
	var err error

	err = d.Engine.Sync2(new(Todo), new(Author))

	if err != nil {
		fmt.Println("Error when syning with database")
		fmt.Println(err)
	}
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

type Author struct {
	Id       int64
	Username string
	Hash     string
	Created  time.Time `xorm:"created"`
}
