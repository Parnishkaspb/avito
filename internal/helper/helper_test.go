package helper

import (
	"github.com/Parnishkaspb/avito/internal/models"
	"reflect"
	"testing"
)

func TestParseMembers_Negative(t *testing.T) {
	tests := []struct {
		name    string
		teamID  string
		members []models.RequestMembers
		want    [][]interface{}
	}{
		{
			name:    "empty members slice",
			teamID:  "team1",
			members: []models.RequestMembers{},
			want:    nil,
		},
		{
			name:    "nil members slice",
			teamID:  "team1",
			members: nil,
			want:    nil,
		},
		{
			name:   "empty teamID",
			teamID: "",
			members: []models.RequestMembers{
				{UserID: "u1"},
			},
			want: nil,
		},
		{
			name:   "member with empty fields",
			teamID: "team1",
			members: []models.RequestMembers{
				{UserID: ""},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMembers(tt.teamID, tt.members)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}
