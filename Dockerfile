FROM alpine:latest
COPY ./bin/IssueTracker /app/IssueTracker
COPY ./config.yaml /app/config.yaml
ENTRYPOINT ["/app/IssueTracker", "/app/config.yaml"]