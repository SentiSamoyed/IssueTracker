package model

import "time"

type Repo struct {
	Id       int64 `gorm:"primaryKey"`
	Owner    string
	Name     string
	FullName string
	HtmlUrl  string
}

type Issue struct {
	Id           int64 `gorm:"primaryKey"`
	RepoFullName string
	IssueNumber  int
	Title        string
	State        string
	HtmlUrl      string
	Author       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Body         string
	Comments     int
}
