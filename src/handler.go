package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func RepoLoadRequestHandler(writer http.ResponseWriter, request *http.Request) {
	if "POST" != request.Method {
		// POST only
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	path := request.URL.Path
	path = strings.TrimPrefix(path, "/repo/")
	ss := strings.Split(path, "/")
	if len(ss) != 2 {
		log.Printf("Bad request: %v\n", path)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	owner, repo := ss[0], ss[1]
	fullName := fmt.Sprintf("%s/%s", owner, repo)

	result := TrackerSubmit(fullName)

	if _, err := writer.Write([]byte(fmt.Sprintf("%v", result))); err != nil {
		log.Printf("Error: Failed to respond %v to %v: %v\n", result, fullName, err.Error())
	}
}
