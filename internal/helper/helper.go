package helper

import (
	"github.com/Parnishkaspb/avito/internal/models"
	"math/rand"
	"time"
)

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
		})
	}

	return rows
}

func PickRandomTeamMates(all []string, n int) []string {
	if len(all) <= n {
		return all
	}

	rand.Seed(time.Now().UnixNano())
	picked := make([]string, n)
	indices := rand.Perm(len(all))

	for i := 0; i < n; i++ {
		picked[i] = all[indices[i]]
	}

	return picked
}
