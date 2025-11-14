package helper

import "github.com/Parnishkaspb/avito/internal/models"

func ParseMembers(teamID string, members []models.RequestMembers) [][]interface{} {
	if len(members) == 0 || teamID == "" {
		return nil
	}

	rows := make([][]interface{}, 0, len(members))

	for _, member := range members {
		if member.UserID == "" {
			return nil
		}

		rows = append(rows, []interface{}{
			teamID,
			member.UserID,
			true,
		})
	}

	return rows
}
