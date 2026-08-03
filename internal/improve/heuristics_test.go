package improve_test

import (
	"reflect"
	"sort"
	"testing"

	"aint/internal/improve"
)

func TestDetectSignalCategories(t *testing.T) {
	cases := []struct {
		text string
		want []string
	}{
		{"that's wrong, try again", []string{"correction"}},
		{"no, that's wrong too", []string{"correction"}},
		{"not what I asked for", []string{"correction"}},
		{"please undo that change", []string{"stop-undo"}},
		{"revert this now", []string{"stop-undo"}},
		{"why did you do that", []string{"stop-undo"}},
		{"ugh, seriously?", []string{"frustration-language"}},
		{"I already told you not to do that", []string{"frustration-language"}},
		{"WHY WOULD YOU DO THAT", []string{"shouting"}},
		{"stop!!", []string{"shouting"}},
		{"no.", []string{"terse-negative-reply"}},
		{"wrong.", []string{"terse-negative-reply"}},
		{"please add a test for this function", nil},
		{"This is a CLI tool", nil},
		{"no worries, that's totally fine, thanks!", nil},
	}

	for _, c := range cases {
		got := improve.Detect(c.text)
		if c.want == nil {
			// For nil expectations, use direct comparison
			if got != nil {
				t.Errorf("Detect(%q) = %v, want nil", c.text, got)
			}
		} else {
			// For non-nil expectations, sort and compare
			sort.Strings(got)
			want := append([]string{}, c.want...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Detect(%q) = %v, want %v", c.text, got, want)
			}
		}
	}
}

func TestDetectReturnsNilForNoSignals(t *testing.T) {
	cases := []string{
		"please add a test for this function",
		"This is a CLI tool",
		"no worries, that's totally fine, thanks!",
		"The quick brown fox jumps over the lazy dog",
		"I appreciate your help",
	}

	for _, text := range cases {
		got := improve.Detect(text)
		if got != nil {
			t.Errorf("Detect(%q) returned non-nil %v (type %T), want nil", text, got, got)
		}
	}
}
