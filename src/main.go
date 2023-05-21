package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

var Db *gorm.DB

type Config struct {
	Server struct {
		Addr string
	}
	Datasource struct {
		User     string
		Password string
		Suffix   string
	}
}

func main() {
	var conf Config
	buf, err := os.ReadFile("config.yaml")
	if err == nil {
		err = yaml.Unmarshal(buf, &conf)
		log.Println(conf)
	}
	if err != nil {
		log.Panic(err.Error())
	}

	/* Connect to the database */
	pw := os.Getenv(conf.Datasource.Password)
	dsn := fmt.Sprintf("%v:%v%v", conf.Datasource.User, pw, conf.Datasource.Suffix)
	if Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		log.Panic(err.Error())
	}

	InitTracker()

	/* Handlers */
	http.HandleFunc("/repo/", RepoLoadRequestHandler)

	/* Launch */
	log.Panic(http.ListenAndServe(conf.Server.Addr, nil))
}
