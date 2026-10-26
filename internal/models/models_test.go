package models

import (
	"strings"
	"testing"
)

// sp returns n spaces
func sp(n int) string { return strings.Repeat(" ", n) }

func TestLeaderbaord_String(t *testing.T) {
	tests := []struct {
		name string
		lb   Leaderboard
		want string
	}{
		{"monthly", MonthlyLeaderboard, "MonthlyLeaderboard"},
		{"global", GlobalLeaderboard, "GlobalLeaderboard"},
		{"unknown positive", Leaderboard(88), "Leaderboard(88)"},
		{"unknown negative", Leaderboard(-88), "Leaderboard(-88)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lb.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLeaderboard_DBKind(t *testing.T) {
	tests := []struct {
		name string
		lb   Leaderboard
		want string
	}{
		{"monthly", MonthlyLeaderboard, "monthly"},
		{"global", GlobalLeaderboard, "global"},
		{"unknown positive", Leaderboard(88), "unknown(88)"},
		{"unknown negative", Leaderboard(-88), "unknown(-88)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lb.DBKind(); got != tt.want {
				t.Errorf("DBKind() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLeaderboard(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Leaderboard
		wantErr string // substr of expected err; "" means no err
	}{
		{"monthly", "monthly", MonthlyLeaderboard, ""},
		{"global", "global", GlobalLeaderboard, ""},
		{"empty", "", 0, `unknown leaderboard ""`},
		{"wrong case", "Monthly", 0, `unknown leaderboard "Monthly"`},
		{"surrounding whitespace", " global ", 0, `unknown leaderboard " global "`},
		{"unsupported", "weekly", 0, `unknown leaderboard "weekly"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLeaderboard(tt.input)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ParseLeaderboard(%q) unexpected err: %v", tt.input, err)
				}
			} else {
				errStr := err.Error()
				if !strings.Contains(errStr, tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", errStr, tt.wantErr)
				}
			}

			if got != tt.want {
				t.Errorf("ParseLeaderboard(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLeaderboard_DBKindRoundTrip(t *testing.T) {
	for _, lb := range []Leaderboard{MonthlyLeaderboard, GlobalLeaderboard} {
		t.Run(lb.String(), func(t *testing.T) {
			got, err := ParseLeaderboard(lb.DBKind())

			if err != nil {
				t.Fatalf("ParseLeaderboard(%q) error: %v", lb.DBKind(), err)
			}

			if got != lb {
				t.Errorf("round trip %v, want %v", got, lb)
			}
		})
	}
}

func TestUser_DetailedString_RoundsFloats(t *testing.T) {
	u := User{
		AcceptanceRate: 88.88,
		AverageTime:    1.001,
		AveragePoints:  2.881,
	}
	got := u.DetailedString()

	for _, sub := range []string{
		"(88.88%)",
		"AvgTime: 1.00s",
		"AvgPts: 2.88",
	} {
		if !strings.Contains(got, sub) {
			t.Errorf("DetailedString() = %q, want it to contain %q", got, sub)
		}
	}
}
