package server

type requestTeamAdd struct {
	TeamName string           `json:"team_name"`
	Members  []requestMembers `json:"members"`
}

type requestMembers struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
