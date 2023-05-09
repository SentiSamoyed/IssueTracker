package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v52/github"
)

const (
	LIST_ISSUE_API_FMT    = "https://api.github.com/repos/%s/%s/issues?per_page=%d&page=%d"
	ISSUE_COMMENT_API_FMT = "https://api.github.com/repos/%s/%s/issues/%d/comments?per_page=%d&page=%d"
)

func main() {
	client := github.NewClient(nil)
	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    0,
		},
	}

	issues, _, err := client.Issues.ListByRepo(context.Background(), "apache", "dubbo", opt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v issues in total.", len(issues))

	for _, p := range issues {
		issue := *p

	}
}
