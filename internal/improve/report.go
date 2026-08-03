// internal/improve/report.go
package improve

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// WriteReport renders incidents as markdown to w, grouped by project then
// by session (chronological within a session). date is formatted into the
// report title (e.g. "2026-08-03"). Writing nothing for an empty incidents
// slice lets callers skip creating a report file entirely when there's
// nothing new.
func WriteReport(w io.Writer, incidents []Incident, date string) error {
	if len(incidents) == 0 {
		return nil
	}

	byProject := map[string][]Incident{}
	for _, inc := range incidents {
		byProject[inc.Project] = append(byProject[inc.Project], inc)
	}
	projects := make([]string, 0, len(byProject))
	for p := range byProject {
		projects = append(projects, p)
	}
	sort.Strings(projects)

	failed := 0
	sessions := map[string]bool{}
	for _, inc := range incidents {
		if inc.AnalysisFailed != "" {
			failed++
		}
		sessions[inc.SessionID] = true
	}

	if _, err := fmt.Fprintf(w, "# aint improve report — %s\n\n%d incidents across %d sessions, %d analysis failures.\n\n", date, len(incidents), len(sessions), failed); err != nil {
		return err
	}

	for _, project := range projects {
		if _, err := fmt.Fprintf(w, "## %s\n\n", project); err != nil {
			return err
		}
		if err := writeProjectSessions(w, byProject[project]); err != nil {
			return err
		}
	}
	return nil
}

func writeProjectSessions(w io.Writer, incidents []Incident) error {
	bySession := map[string][]Incident{}
	for _, inc := range incidents {
		bySession[inc.SessionID] = append(bySession[inc.SessionID], inc)
	}
	sessionIDs := make([]string, 0, len(bySession))
	for s := range bySession {
		sessionIDs = append(sessionIDs, s)
	}
	sort.Strings(sessionIDs)

	for _, sessionID := range sessionIDs {
		list := bySession[sessionID]
		sort.Slice(list, func(i, j int) bool { return list[i].Line < list[j].Line })

		if _, err := fmt.Fprintf(w, "### Session %s — %s\n\n", sessionID, list[0].Timestamp.Format("2006-01-02T15:04:05Z")); err != nil {
			return err
		}
		for _, inc := range list {
			if err := writeIncident(w, inc); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeIncident(w io.Writer, inc Incident) error {
	if _, err := fmt.Fprintf(w, "**Signals:** %s\n**User said:**\n> %q\n\n", joinOrNone(inc.Signals), inc.Text); err != nil {
		return err
	}
	if inc.AnalysisFailed != "" {
		_, err := fmt.Fprintf(w, "⚠️ claude analysis unavailable: %s\n\n---\n\n", inc.AnalysisFailed)
		return err
	}
	_, err := fmt.Fprintf(w,
		"**What happened:** %s\n**Root cause:** %s\n**Suggested aint rule:** %s\n**Suggested lint rule:** %s\n**Suggested doc/memory change:** %s\n\n---\n\n",
		inc.Summary, inc.RootCause,
		orNotApplicable(inc.AintRuleSuggestion),
		orNotApplicable(inc.LintRuleSuggestion),
		orNotApplicable(inc.DocMemorySuggestion))
	return err
}

func joinOrNone(signals []string) string {
	if len(signals) == 0 {
		return "none"
	}
	return strings.Join(signals, ", ")
}

func orNotApplicable(s string) string {
	if s == "" {
		return "Not applicable"
	}
	return s
}
