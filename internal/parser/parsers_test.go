package parser

import (
	"math"
	"testing"
	"time"

	"nook/internal/models"
)

const epsilon = 1e-9

const (
	sitePrefix   = "https://starbattle.puzzlebaron.com/profile.php"
	urlSeparator = "?u="
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func buildURL(prefix, sep, uid string) string {
	return prefix + sep + uid
}

func TestUID(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", buildURL("", "", ""), -1},
		{"no separator", buildURL(sitePrefix, "", "88"), -1},
		{"malformed separator", buildURL(sitePrefix, "?=", "88"), -1},
		{"negative uid", buildURL(sitePrefix, urlSeparator, "-1"), -1},
		{"zero uid", buildURL(sitePrefix, urlSeparator, "0"), 0},
		{"positive uid one digit uid", buildURL(sitePrefix, urlSeparator, "8"), 8},
		{"positive uid two digit uid", buildURL(sitePrefix, urlSeparator, "19"), 19},
		{"positive uid three digit uid", buildURL(sitePrefix, urlSeparator, "444"), 444},
		{"positive uid four digit uid", buildURL(sitePrefix, urlSeparator, "8888"), 8888},
		{"positive uid five digit uid", buildURL(sitePrefix, urlSeparator, "10305"), 10305},
		{"positive uid six digit uid", buildURL(sitePrefix, urlSeparator, "111111"), 111111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := UID(tt.input)
			if got != tt.want {
				t.Errorf("UID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUID_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", buildURL(sitePrefix, urlSeparator, "10305"), false},
		{"empty", "", true},
		{"no separator", buildURL(sitePrefix, "", "88"), true},
		{"malformed separator", buildURL(sitePrefix, "?=", "88"), true},
		{"separator but no uid", buildURL(sitePrefix, urlSeparator, ""), true},
		{"non-numeric uid", buildURL(sitePrefix, urlSeparator, "?u=abc"), true},
		{"uid with trailing params", buildURL(sitePrefix, urlSeparator, "42&x=1"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestRank(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"one digit", "1.", 1, false},
		{"two digits", "42.", 42, false},
		{"three digits", "123.", 123, false},
		{"four digits", "9999.", 9999, false},
		{"zero", "0.", 0, false},
		{"with trailing text", "7. username", 7, false},
		{"multiple dots splits at first", "5.3.1", 5, false},
		{"empty", "", -1, true},
		{"no dot", "12", -1, true},
		{"only dot", ".", -1, true},
		{"non-numeric", "abc.", -1, true},
		{"leading space", " 5.", -1, true},
		{"decimal-like", "1,000.", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Rank(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Rank(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				if got != -1 {
					t.Errorf("Rank(%q) = %v on error, want -1", tt.input, got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("Rank(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDate(t *testing.T) {
	ref := time.Date(2024, time.April, 19, 0, 0, 0, 0, models.SiteLocation)

	layouts := []struct {
		name   string
		layout string
	}{
		{"month layout", models.MonthDateLayout},
		{"global layout", models.GlobalDateLayout},
	}

	for _, l := range layouts {
		formatted := ref.Format(l.layout)

		t.Run(l.name+"/valid", func(t *testing.T) {
			got, err := Date(formatted)
			if err != nil {
				t.Fatalf("Date(%q) unexpected error: %v", formatted, err)
			}
			if got.Format(l.layout) != formatted {
				t.Errorf("Date(%q) round trip = %q", formatted, got.Format(l.layout))
			}
		})

		t.Run(l.name+"/uses site location", func(t *testing.T) {
			got, err := Date(formatted)
			if err != nil {
				t.Fatalf("Date(%q) unexpected error: %v", formatted, err)
			}
			if got.Location() != models.SiteLocation {
				t.Errorf("Date(%q) location = %v, want %v",
					formatted, got.Location(), models.SiteLocation)
			}
		})

		t.Run(l.name+"/trims whitespace", func(t *testing.T) {
			padded := "  " + formatted + "\n\t"
			got, err := Date(padded)
			if err != nil {
				t.Fatalf("Date(%q) unexpected error: %v", padded, err)
			}
			if got.Format(l.layout) != formatted {
				t.Errorf("Date(%q) = %q, want %q", padded, got.Format(l.layout), formatted)
			}
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"garbage", "not a date"},
		{"impossible date", "2024-13-45"},
		{"number only", "12345"},
	}

	for _, tt := range invalid {
		t.Run("invalid/"+tt.name, func(t *testing.T) {
			got, err := Date(tt.input)
			if err == nil {
				t.Fatalf("Date(%q) = %v, want error", tt.input, got)
			}
			if !got.IsZero() {
				t.Errorf("Date(%q) = %v on error, want zero time", tt.input, got)
			}
		})
	}
}

func TestSolved(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"plain number", "5", 5, false},
		{"zero", "0", 0, false},
		{"thousands separator", "1,234", 1234, false},
		{"multiple separators", "1,234,567", 1234567, false},
		{"with trailing text", "88 solved", 88, false},
		{"separator with trailing text", "5,888 puzzles", 5888, false},
		{"leading whitespace", "   42 puzzles", 42, false},
		{"non-numeric", "abc", 0, true},
		{"non-numeric first field", "abc 5", 0, true},
		{"decimal", "1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Solved(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Solved(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Solved(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAttempted(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"plain number", "10", 10, false},
		{"zero", "0", 0, false},
		{"thousands separator", "12,345", 12345, false},
		{"with trailing text", "300 attempted", 300, false},
		{"non-numeric", "n/a", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Attempted(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Attempted(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Attempted(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSolvedAttempted(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSolved    int
		wantAttempted int
		wantErr       bool
	}{
		{"simple", "5/10", 5, 10, false},
		{"zeros", "0/0", 0, 0, false},
		{"equal", "7/7", 7, 7, false},
		{"large", "12345/67890", 12345, 67890, false},
		{"empty", "", -1, -1, true},
		{"no separator", "510", -1, -1, true},
		{"wrong separator", "5-10", -1, -1, true},
		{"bad solved", "abc/10", -1, -1, true},
		{"empty solved", "/10", -1, -1, true},
		{"bad attempted", "5/abc", 5, -1, true},
		{"empty attempted", "5/", 5, -1, true},
		{"spaces not trimmed", " 5/10", -1, -1, true},
		{"thousands separators not handled", "1,234/5,678", -1, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solved, attempted, err := SolveAttempted(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SolveAttempted(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if solved != tt.wantSolved || attempted != tt.wantAttempted {
				t.Errorf("SolveAttempted(%q) = (%v, %v), want (%v, %v)",
					tt.input, solved, attempted, tt.wantSolved, tt.wantAttempted)
			}
		})
	}
}

func TestAcceptanceRate(t *testing.T) {
	tests := []struct {
		name      string
		solved    int
		attempted int
		want      float64
	}{
		{"zero attempted", 5, 0, 0},
		{"both zero", 0, 0, 0},
		{"zero solved", 0, 5, 0},
		{"all solved", 1, 1, 100},
		{"half", 1, 2, 50},
		{"quarter", 1, 4, 25},
		{"three quarters", 3, 4, 75},
		{"one eighth", 1, 8, 12.5},
		{"truncates down 1/3", 1, 3, 33.33},
		{"truncates down 2/3", 2, 3, 66.66},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AcceptanceRate(tt.solved, tt.attempted)
			if !approxEqual(got, tt.want) {
				t.Errorf("AcceptanceRate(%d, %d) = %v, want %v",
					tt.solved, tt.attempted, got, tt.want)
			}
		})
	}
}

func TestAvgTime_Global(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"seconds only", "45", 45, false},
		{"zero", "0", 0, false},
		{"minutes and seconds", "2:35", 155, false},
		{"hours minutes seconds", "1:02:37", 3757, false},
		{"zero padded", "0:00", 0, false},
		{"exactly one hour", "1:00:00", 3600, false},
		{"surrounding whitespace", "  1:30  ", 90, false},
		{"empty", "", 0, true},
		{"too many parts", "1:2:3:4", 0, true},
		{"non-numeric", "a:b", 0, true},
		{"negative minutes", "-1:30", 0, true},
		{"negative seconds", "1:-30", 0, true},
		{"decimal", "1.5", 0, true},
		{"trailing colon", "1:", 0, true},
		{"leading colon", ":30", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AvgTimeGlobal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AvgTimeGlobal(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !approxEqual(got, tt.want) {
				t.Errorf("AvgTimeGlobal(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAvgTime_Monthly(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"no unit suffix", "44.3", 44.3, false},
		{"integer", "10 sec.", 10, false},
		{"decimal", "2.5 sec.", 2.5, false},
		{"zero", "0 sec.", 0, false},
		{"leading whitespace", "   4.5 sec.", 4.5, false},
		{"non-numeric", "abc", 0, true},
		{"non-numeric first field", "abc 3.5", 0, true},
		{"thousands separator not handled", "1,234.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AvgTimeMonthly(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AvgTimeMonthly(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !approxEqual(got, tt.want) {
				t.Errorf("AvgTimeMonthly(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAvgPoints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"no unit", "19", 19, false},
		{"integer", "100 pts.", 100, false},
		{"decimal", "123.45 pts.", 123.45, false},
		{"zero", "0 pts.", 0, false},
		{"leading whitespace", "  9.25 points", 9.25, false},
		{"non-numeric", "n/a", 0, true},
		{"thousands separator not handled", "1,000.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AvgPoints(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AvgPoints(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !approxEqual(got, tt.want) {
				t.Errorf("AvgPoints(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAvgPointsGlobal(t *testing.T) {
	// NOTE: this documents current behavior: integer division truncates.
	tests := []struct {
		name        string
		totalPoints int
		solved      int
		want        float64
	}{
		{"exact division", 100, 10, 10},
		{"zero points", 0, 5, 0},
		{"truncates", 10, 4, 2},
		{"solved one", 77, 1, 77},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AvgPointsGlobal(tt.totalPoints, tt.solved)
			if err != nil {
				t.Fatalf("AvgPointsGlobal(%d, %d) unexpected error: %v",
					tt.totalPoints, tt.solved, err)
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("AvgPointsGlobal(%d, %d) = %v, want %v",
					tt.totalPoints, tt.solved, got, tt.want)
			}
		})
	}
}

func TestTotalPoints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"zero", "0", 0, false},
		{"small", "12", 12, false},
		{"thousands separator", "1,234", 1234, false},
		{"multiple separators", "1,234,567", 1234567, false},
		{"empty", "", 0, true},
		{"non-numeric", "abc", 0, true},
		{"leading space not trimmed", " 5", 0, true},
		{"decimal", "1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TotalPoints(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("TotalPoints(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("TotalPoints(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"zero", "0", 0, false},
		{"small", "42", 42, false},
		{"thousands separator", "5,888", 5888, false},
		{"multiple separators", "1,234,567", 1234567, false},
		{"surrounding whitespace", "  1,234  ", 1234, false},
		{"empty", "", 0, true},
		{"whitespace only", "   ", 0, true},
		{"non-numeric", "abc", 0, true},
		{"decimal", "1.5", 0, true},
		{"inner whitespace", "1 234", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Count(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Count(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Count(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAvgTimeClock(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"zero", "0:00", 0, false},
		{"thirty seconds", "0:30", 30, false},
		{"one minute", "1:00", 60, false},
		{"rounds to two decimals", "2:35", 155, false},
		{"hours", "1:02:37", 3757, false},
		{"exactly one hour", "1:00:00", 3600, false},
		{"surrounding whitespace", "  2:30  ", 150, false},
		{"empty", "", -1, true},
		{"single part", "5", -1, true},
		{"too many parts", "1:2:3:4", -1, true},
		{"non-numeric", "a:30", -1, true},
		{"negative minutes", "-1:30", -1, true},
		{"negative seconds", "1:-30", -1, true},
		{"decimal", "1.5:30", -1, true},
		{"trailing colon", "1:", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AvgTimeClock(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AvgTimeClock(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !approxEqual(got, tt.want) {
				t.Errorf("AvgTimeClock(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAvgPointsDivide(t *testing.T) {
	tests := []struct {
		name   string
		total  int
		solved int
		want   float64
	}{
		{"zero solved", 5, 0, 0},
		{"zero total", 0, 5, 0},
		{"exact", 10, 4, 2.5},
		{"rounds down", 100, 3, 33.33},
		{"rounds up", 200, 3, 66.67},
		{"one third", 1, 3, 0.33},
		{"whole number", 100, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AvgPointsDivide(tt.total, tt.solved)
			if !approxEqual(got, tt.want) {
				t.Errorf("AvgPointsDivide(%d, %d) = %v, want %v",
					tt.total, tt.solved, got, tt.want)
			}
		})
	}
}
