package feed

import "testing"

func TestCategoryFromLabels(t *testing.T) {
	cases := []struct {
		labels []string
		want   string
	}{
		{[]string{"kind/bug"}, "bug"},
		{[]string{"Documentation"}, "documentation"},
		{[]string{"docs", "help wanted"}, "documentation"},
		{[]string{"refactoring"}, "refactoring"},
		{[]string{"code cleanup"}, "refactoring"},
		{[]string{"Tech Debt"}, "refactoring"},
		{[]string{"enhancement"}, "feature"},
		{nil, "feature"},
		{[]string{"docs", "bug"}, "bug"},
	}
	for _, c := range cases {
		if got := categoryFromLabels(c.labels); got != c.want {
			t.Errorf("categoryFromLabels(%v) = %q, want %q", c.labels, got, c.want)
		}
	}
}

func TestDifficultyFromLabels(t *testing.T) {
	cases := []struct {
		labels []string
		want   string
	}{
		{[]string{"good first issue"}, "easy"},
		{[]string{"E-easy"}, "easy"},
		{[]string{"beginner friendly"}, "easy"},
		{[]string{"hard"}, "hard"},
		{[]string{"complexity: complex"}, "hard"},
		{[]string{"help wanted"}, "medium"},
		{nil, "medium"},
	}
	for _, c := range cases {
		if got := difficultyFromLabels(c.labels); got != c.want {
			t.Errorf("difficultyFromLabels(%v) = %q, want %q", c.labels, got, c.want)
		}
	}
}
