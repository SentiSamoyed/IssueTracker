package model

import "time"

type Repo struct {
	Id       *int64 `gorm:"primaryKey"`
	Owner    *string
	Name     *string
	FullName *string
	HtmlUrl  *string
}

type Issue struct {
	Id           *int64 `gorm:"primaryKey"`
	RepoFullName *string

	IssueNumber *int
	Title       *string
	State       *string
	HtmlUrl     *string
	Author      *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Body        *string
	Comments    *int
}

type Comment struct {
	Id           *int64 `gorm:"primaryKey"`
	RepoFullName *string
	IssueNumber  *int
	HtmlUrl      *string
	Author       *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Body         *string
}

type Release struct {
	Id           *int64 `gorm:"primaryKey"`
	RepoFullName *string
	TagName      *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}
