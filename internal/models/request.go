package models

type RequestTeamAdd struct {
	TeamName string           `json:"team_name"`
	Members  []RequestMembers `json:"members"`
}

type RequestMembers struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
