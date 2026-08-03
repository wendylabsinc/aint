// Package improve provides tools for analyzing Claude Code session transcripts
// to identify user frustration, disagreement, and areas for improvement.
package improve

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HumanMessage is a single human-typed chat line extracted from a Claude
// Code session transcript.
type HumanMessage struct {
	SessionFile string
	Line        int // 1-based line number within SessionFile
	Timestamp   time.Time
	Project     string // cwd recorded on the transcript line
	SessionID   string
	Text        string
}

// AssistantTurn is a single assistant-authored line extracted from a
// session transcript, with its text and tool_use blocks flattened into one
// summary string in original block order.
type AssistantTurn struct {
	Line      int
	Timestamp time.Time
	Rendered  string
}

type rawTranscriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	Origin    struct {
		Kind string `json:"kind"`
	} `json:"origin"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// FindSessionFiles returns every *.jsonl file under dir, sorted
// lexicographically for deterministic processing order.
func FindSessionFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// ParseSessionFile reads path, skipping the first startLine lines (0 means
// read from the start), and returns every human-typed message and every
// assistant turn found from there on, plus the total number of lines in
// the file. Lines that fail to parse as JSON are skipped rather than
// causing an error, since a session file being actively appended to can
// have a truncated trailing line. For user messages with string content,
// a message is treated as human if origin.kind is either "human" or absent
// (empty); messages with explicit non-human kinds are excluded.
func ParseSessionFile(path string, startLine int) ([]HumanMessage, []AssistantTurn, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()

	var humans []HumanMessage
	var assistants []AssistantTurn
	lineNum := 0

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		lineNum++
		if lineNum <= startLine {
			continue
		}

		var raw rawTranscriptLine
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}
		ts, _ := time.Parse(time.RFC3339, raw.Timestamp)

		switch raw.Type {
		case "user":
			var text string
			if err := json.Unmarshal(raw.Message.Content, &text); err != nil {
				continue // array content: a tool_result being echoed back, not human-typed
			}
			// Treat as human if origin.kind is "human" or absent (empty).
			// Exclude only explicit non-human kinds (e.g. "task-notification", "coordinator").
			if raw.Origin.Kind != "human" && raw.Origin.Kind != "" {
				continue
			}
			humans = append(humans, HumanMessage{
				SessionFile: path,
				Line:        lineNum,
				Timestamp:   ts,
				Project:     raw.CWD,
				SessionID:   raw.SessionID,
				Text:        text,
			})
		case "assistant":
			var blocks []contentBlock
			if err := json.Unmarshal(raw.Message.Content, &blocks); err != nil {
				continue
			}
			rendered := renderBlocks(blocks)
			if rendered == "" {
				continue
			}
			assistants = append(assistants, AssistantTurn{
				Line:      lineNum,
				Timestamp: ts,
				Rendered:  rendered,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return humans, assistants, lineNum, err
	}
	return humans, assistants, lineNum, nil
}

func renderBlocks(blocks []contentBlock) string {
	var parts []string
	for _, blk := range blocks {
		switch blk.Type {
		case "text":
			if t := strings.TrimSpace(blk.Text); t != "" {
				parts = append(parts, t)
			}
		case "tool_use":
			parts = append(parts, blk.Name+"("+summarizeInput(blk.Input)+")")
		}
	}
	return strings.Join(parts, "\n")
}

func summarizeInput(raw json.RawMessage) string {
	s := strings.Join(strings.Fields(string(raw)), " ")
	const max = 160
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
