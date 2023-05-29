package main

import (
	"context"
	"fmt"
	"github.com/SentiSamoyed/IssueTracker/model"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"strings"
	"time"
)

type Answer int

const (
	MainChanCap    int = 100
	CommentChanCap int = 100
)

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
var taskQueue chan string

func InitTracker() {
	reqChan = make(chan TrackerRequest, 50)
	resChan = make(chan TrackerResult, 50)
	taskQueue = make(chan string, 100)
	go mainLoop()
	go handleRequest()
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
				taskQueue <- fullName
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

func handleRequest() {
	for {
		fullName := <-taskQueue
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
}

func scrapeRepo(fullName string) (ans Answer, err error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: os.Getenv("GH_TOKEN"),
		},
	)
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))

	ss := strings.Split(fullName, "/")
	owner, repo := ss[0], ss[1]

	/* Check the database first */
	repoPo := model.Repo{}
	result := Db.Table("repo").Where("full_name = ?", fullName).Take(&repoPo)
	existed := result.Error == nil
	if existed {
		repoLog(fullName, "Has been stored in the database. Now trying to update...")
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

	if !existed {
		repoPo = model.Repo{
			Id:       repository.ID,
			Owner:    repository.Owner.Login,
			Name:     repository.Name,
			FullName: repository.FullName,
			HtmlUrl:  repository.HTMLURL,
		}

		if result := tx.Table("repo").Create(&repoPo); result.Error != nil {
			return Failed, result.Error
		}

		repoLog(fullName, "Repo written to the database")
	}

	var lastIssue model.Issue
	var sinceIssue *time.Time
	if tx.Table("issue").Where("repo_full_name = ?", fullName).Order("created_at desc").Limit(1).Find(&lastIssue).RowsAffected == 1 {
		sinceIssue = lastIssue.CreatedAt
		repoLog(fullName, "Getting issues since %v", sinceIssue)
	}

	/* Get Issues and Comments */
	ch := make(chan interface{}, MainChanCap)
	done := 0
	sum := int64(0)
	go getReleases(client, fullName, repository, ch)
	go getIssues(client, fullName, repository, sinceIssue, ch)

	finished := 3 // Release + Issue + Comment(Created by getIssues)
	for done < finished {
		res := <-ch
		delta := int64(0)
		switch r := res.(type) {
		case []*model.Release:
			result := tx.Table("release").Clauses(clause.OnConflict{DoNothing: true}).Create(r)
			delta = result.RowsAffected
			err = result.Error
		case []*model.Issue:
			result := tx.Table("issue").Clauses(clause.OnConflict{DoNothing: true}).Create(r)
			delta = result.RowsAffected
			err = result.Error
		case []*model.Comment:
			result := tx.Table("comment").Clauses(clause.OnConflict{DoNothing: true}).Create(r)
			delta = result.RowsAffected
			err = result.Error
		case error:
			err = r
			done++
		case Answer:
			done++
		default:
			err = fmt.Errorf("unexpected type: %v", r)
		}

		if err != nil {
			repoLog(fullName, "Error: "+err.Error())
		}

		sum += delta
	}

	repoLog(fullName, "Releases, Issues, and comments written to the database.")
	repoLog(fullName, "Written %v rows in total.", sum)

	return Done, nil
}

func getReleases(client *github.Client, fullName string, repo *github.Repository, ch chan interface{}) {
	opts := github.ListOptions{
		Page:    0,
		PerPage: 100,
	}

	for {
		owner, name := *repo.Owner.Login, *repo.Name
		releases, resp, err := client.Repositories.ListReleases(context.Background(), owner, name, &opts)
		if err != nil {
			ch <- err
			return
		}

		size := len(releases)
		releasePos := make([]*model.Release, 0, size)
		for _, release := range releases {
			releasePos = append(releasePos, &model.Release{
				Id:           release.ID,
				RepoFullName: &fullName,
				TagName:      release.TagName,
				CreatedAt:    &release.CreatedAt.Time,
			})
		}
		ch <- releasePos

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	repoLog(fullName, "Releases done")
	ch <- Done
}

func getIssues(client *github.Client, fullName string, repo *github.Repository, since *time.Time, ch chan interface{}) {
	opts := github.IssueListByRepoOptions{
		State:     "all",
		Sort:      "created",
		Direction: "asc",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	if since != nil {
		opts.Since = *since
	}

	commentCh := make(chan interface{}, CommentChanCap)
	go getComments(client, fullName, repo, since, commentCh, ch)
	defer func() {
		commentCh <- Done
	}()

	sum := 0
	milestone := 0
	for {
		owner, name := *repo.Owner.Login, *repo.Name
		issues, resp, err := client.Issues.ListByRepo(context.Background(), owner, name, &opts)

		if err != nil {
			ch <- err
			return
		}

		sum += len(issues)
		if sum >= milestone*100 {
			repoLog(fullName, "%v issues saved", sum)
			milestone = (sum + 99) / 100
		}
		dealWithIssues(issues, fullName, commentCh, ch)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	repoLog(fullName, "Issues done")
	ch <- Done
}

func dealWithIssues(issues []*github.Issue, fullName string, commentCh chan interface{}, ch chan interface{}) {
	size := len(issues)
	if size > 0 {
		issuePos := make([]*model.Issue, 0, size)
		for _, issue := range issues {
			if issue.PullRequestLinks != nil {
				continue
			}
			issuePos = append(issuePos, &model.Issue{
				Id:           issue.ID,
				RepoFullName: &fullName,
				IssueNumber:  issue.Number,
				Title:        issue.Title,
				State:        issue.State,
				HtmlUrl:      issue.HTMLURL,
				Author:       issue.User.Login,
				CreatedAt:    &issue.CreatedAt.Time,
				UpdatedAt:    &issue.UpdatedAt.Time,
				Body:         issue.Body,
				Comments:     issue.Comments,
			})
			if *issue.Comments > 0 {
				commentCh <- *issue.Number
			}
		}
		// Double check
		if len(issuePos) > 0 {
			ch <- issuePos
		}
	}
}

func getComments(client *github.Client, fullName string, repo *github.Repository, since *time.Time, in chan interface{}, ch chan interface{}) {
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

	defer func() {
		ch <- Done
	}()

	for {
		var issueNumber int
		switch r := (<-in).(type) {
		case Answer:
			ch <- Done
			return
		case int:
			issueNumber = r
		}

		var outerErr error
	outer:
		for {
			owner, name := *repo.Owner.Login, *repo.Name
			comments, resp, err := client.Issues.ListComments(context.Background(), owner, name, issueNumber, &opts)
			if err != nil {
				outerErr = err
				break outer
			}

			size := len(comments)
			if size > 0 {
				commentPos := make([]*model.Comment, size, size)
				for i, c := range comments {
					if err != nil {
						outerErr = err
						break outer
					}
					commentPos[i] = &model.Comment{
						Id:           c.ID,
						RepoFullName: &fullName,
						IssueNumber:  &issueNumber,
						HtmlUrl:      c.HTMLURL,
						Author:       c.User.Login,
						CreatedAt:    &c.CreatedAt.Time,
						UpdatedAt:    &c.UpdatedAt.Time,
						Body:         c.Body,
					}
				}
				ch <- commentPos
			}

			if resp.NextPage == 0 {
				break
			}

			opts.Page = resp.NextPage
		}

		if outerErr != nil {
			ch <- outerErr
			for {
				switch (<-in).(type) {
				case Answer:
					return
				default:
				}
			}
		}
	}
}
