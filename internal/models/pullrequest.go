package models

import "time"

type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          string
}

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	MergedAt          *time.Time
}
