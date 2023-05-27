#!/usr/bin/env bash

cd ./src || exit
http_proxy=http://127.0.0.1:7890 https_proxy=http://127.0.0.1:7890 GOOS=linux GOARCH=amd64 go build -o ../bin/IssueTracker
