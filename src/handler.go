package main

import (
	"context"
	"google.golang.org/appengine/log"
	"net/http"
	"strings"
)

func RepoLoadRequestHandler(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	ss := strings.Split(path, "/")
	if len(ss) != 3 {
		log.Infof(context.Background(), "Bad request: %v", path)
		writer.WriteHeader(400)
		return
	}

	owner, repo := ss[1], ss[2]

}
