package models

type TeamMembers struct {
	ID       string
	UserID   string
	TeamID   string
	IsActive bool
	IsAdmin  bool
}
