#!/usr/bin/env bash

cd ./src || exit
GOOS=linux GOARCH=amd64 go build -o ../bin/IssueTracker
