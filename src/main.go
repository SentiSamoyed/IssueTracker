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
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %v <path/to/config.yaml>\n", os.Args[0])
		os.Exit(1)
	}
	var conf Config
	buf, err := os.ReadFile(os.Args[1])
	if err == nil {
		err = yaml.Unmarshal(buf, &conf)
		log.Printf("> Config read: %v\n", conf)
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
	log.Println("> Listening on " + conf.Server.Addr)
	log.Panic(http.ListenAndServe(conf.Server.Addr, nil))
}
