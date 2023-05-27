FROM golang:1.19-alpine AS build
LABEL authors="nintensaga"

COPY ./src/ /src/
WORKDIR /src/

RUN go build -o IssueTracker

FROM alpine:latest
COPY --from=build /src/IssueTracker /app/IssueTracker
COPY ./config.yaml /app/config.yaml
ENTRYPOINT ["/app/IssueTracker", "/app/config.yaml"]