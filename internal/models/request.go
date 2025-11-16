package models

type RequestMembers struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type UserActive struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type PullRequestCreateRequest struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}
