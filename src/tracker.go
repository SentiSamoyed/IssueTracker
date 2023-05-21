package main

import (
	"context"
	"fmt"
	"github.com/SentiSamoyed/IssueTracker/src/model"
	"github.com/google/go-github/v52/github"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Answer int

const (
	Undone   Answer = 0
	Done     Answer = 1
	NotExist Answer = 2
	NoIssue  Answer = 3
	Failed   Answer = iota
)

type TrackerRequest struct {
	FullName string
	Chan     chan Answer
}

type TrackerResult struct {
	FullName string
	Answer   Answer
	Err      error
}

var reqChan chan TrackerRequest
var resChan chan TrackerResult

func InitTracker() {
	reqChan = make(chan TrackerRequest, 50)
	resChan = make(chan TrackerResult, 50)
	go mainLoop()
}

func TrackerSubmit(fullName string) Answer {
	ch := make(chan Answer, 1)
	reqChan <- TrackerRequest{
		FullName: fullName,
		Chan:     ch,
	}

	return <-ch
}

func mainLoop() {
	tasks := make(map[string]Answer, 0)

	for {
		select {
		case req := <-reqChan:
			fullName := req.FullName
			ch := req.Chan
			if ans, ex := tasks[fullName]; ex {
				// The task has been done before
				ch <- ans
			} else {
				tasks[fullName] = Undone
				go handleRequest(fullName)
				ch <- Undone
			}

			break
		case res := <-resChan:
			if res.Answer != Failed {
				tasks[res.FullName] = res.Answer
			} else {
				log.Printf("Error on %v: %v\n", res.FullName, res.Err.Error())
			}

			break
		default:
			break
		}
	}
}

func repoLog(fullName string, fmt string, value ...interface{}) {
	log.Printf("["+fullName+"]\t"+fmt+"\n", value...)
}

func handleRequest(fullName string) {
	repoLog(fullName, "Received request")
	ans, err := scrapeRepo(fullName)
	if err != nil {
		repoLog(fullName, "Error: %v", err.Error())
	} else {
		repoLog(fullName, "Done")
	}

	resChan <- TrackerResult{
		FullName: fullName,
		Answer:   ans,
		Err:      err,
	}
}

func scrapeRepo(fullName string) (ans Answer, err error) {
	client := github.NewClient(nil)
	ss := strings.Split(fullName, "/")
	owner, repo := ss[0], ss[1]

	/* Check the database first */
	repoPo := model.Repo{}
	result := Db.Table("repo").Where("full_name = ?", fullName).Take(&repoPo)
	if result.Error == nil {
		repoLog(fullName, "Has been stored in the database")
		return Done, nil
	}

	/* Get the repository */
	repository, _, err := client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return NotExist, err
	} else if !*repository.HasIssues {
		return NoIssue, fmt.Errorf("this repository has no issues ")
	}
	repoLog(fullName, "Got "+*repository.HTMLURL)

	tx := Db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// TODO: issues and comments can retrieve concurrently, then written with this single thread.

	return Done, nil
}

func saveIssues(client *github.Client, fullName string, repo *github.Repository, since *time.Time, waitGroup *sync.WaitGroup, ch chan interface{}) {
	opts := github.IssueListByRepoOptions{
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	if since != nil {
		opts.Since = *since
	}

	var issues []*github.Issue
	var err error
	for i := 0; i > 0 && len(issues) != 0; i++ {
		opts.ListOptions.Page = i
		owner, name := *repo.Owner.Name, *repo.Name
		issues, _, err = client.Issues.ListByRepo(context.Background(), owner, name, &opts)
		if err != nil {
			ch <- err
			return
		}

		size := len(issues)
		issuePos := make([]model.Issue, size, size)
		for i, issue := range issues {
			issuePos[i] = model.Issue{
				Id:           *issue.ID,
				RepoFullName: fullName,
				IssueNumber:  *issue.Number,
				Title:        *issue.Title,
				State:        *issue.State,
				HtmlUrl:      *issue.HTMLURL,
				Author:       *issue.User.Name,
				CreatedAt:    issue.CreatedAt.Time,
				UpdatedAt:    issue.UpdatedAt.Time,
				Body:         *issue.Body,
				Comments:     *issue.Comments,
			}
		}
		ch <- issuePos
	}

	waitGroup.Done()
}

func getComments(client *github.Client, fullName string, repo *github.Repository, since *time.Time, waitGroup *sync.WaitGroup, ch chan interface{}) {
	sSort := "created"
	sDir := "desc"

	opts := github.IssueListCommentsOptions{
		Sort:      &sSort,
		Direction: &sDir,
		Since:     since,
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	var comments []*github.IssueComment
	var err error
	for i := 0; i > 0 && len(comments) != 0; i++ {
		opts.ListOptions.Page = i
		owner, name := *repo.Owner.Name, *repo.Name
		comments, _, err = client.Issues.ListComments(context.Background(), owner, name, 0, &opts)
		if err != nil {
			ch <- err
			return
		}

		size := len(comments)
		commentPos := make([]model.Comment, size, size)
		for i, c := range comments {
			ss := strings.Split(*c.IssueURL, "/")
			issueNum, err := strconv.Atoi(ss[len(ss)-1])
			if err != nil {
				ch <- err
				return
			}
			commentPos[i] = model.Comment{
				Id:           *c.ID,
				RepoFullName: fullName,
				IssueNumber:  issueNum,
				HtmlUrl:      *c.HTMLURL,
				Author:       *c.User.Name,
				CreatedAt:    c.CreatedAt.Time,
				UpdatedAt:    c.UpdatedAt.Time,
				Body:         *c.Body,
			}
		}
		ch <- commentPos
	}

	waitGroup.Done()
}
