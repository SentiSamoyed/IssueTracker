package main

import (
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
		Dsn string
	}
}

func main() {
	var conf Config
	buf, err := os.ReadFile("config.yaml")
	if err == nil {
		err = yaml.Unmarshal(buf, &conf)
	}
	if err != nil {
		log.Panic(err.Error())
	}

	/* Connect to the database */
	if Db, err = gorm.Open(mysql.Open(conf.Datasource.Dsn), &gorm.Config{}); err != nil {
		log.Panic(err.Error())
	}

	/* Handlers */
	http.HandleFunc("/repo/", RepoLoadRequestHandler)

	/* Launch */
	log.Panic(http.ListenAndServe(conf.Server.Addr, nil))
}
