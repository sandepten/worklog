package notes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Parser handles reading and parsing markdown notes
type Parser struct {
	notesDir      string
	workplaceName string
}

// NewParser creates a new note parser
func NewParser(notesDir, workplaceName string) *Parser {
	return &Parser{
		notesDir:      notesDir,
		workplaceName: workplaceName,
	}
}

// ParseFile reads and parses a markdown note file
func (p *Parser) ParseFile(filePath string) (*Note, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	note := &Note{
		FilePath:      filePath,
		Aliases:       []string{},
		Tags:          []string{},
		PendingWork:   []WorkItem{},
		CompletedWork: []WorkItem{},
	}

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	inPendingSection := false
	inCompletedSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Handle frontmatter
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				continue
			}
		}

		if inFrontmatter {
			p.parseFrontmatterLine(line, note)
			continue
		}

		// Handle title
		if strings.HasPrefix(line, "# ") {
			note.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		// Handle summary fields
		if strings.HasPrefix(line, "summary::") {
			note.Summary = strings.TrimSpace(strings.TrimPrefix(line, "summary::"))
			continue
		}

		if strings.HasPrefix(line, "yesterday's summary::") {
			note.YesterdaySummary = strings.TrimSpace(strings.TrimPrefix(line, "yesterday's summary::"))
			continue
		}

		// Handle sections
		if strings.HasPrefix(line, "## Pending Work") {
			inPendingSection = true
			inCompletedSection = false
			continue
		}

		if strings.HasPrefix(line, "## Work Completed") {
			inPendingSection = false
			inCompletedSection = true
			continue
		}

		// Handle work items
		if inPendingSection {
			if item := p.parseWorkItem(line); item != nil {
				note.PendingWork = append(note.PendingWork, *item)
			}
		}

		if inCompletedSection {
			if item := p.parseWorkItem(line); item != nil {
				note.CompletedWork = append(note.CompletedWork, *item)
			}
		}
	}

	return note, scanner.Err()
}

// parseFrontmatterLine parses a single frontmatter line
func (p *Parser) parseFrontmatterLine(line string, note *Note) {
	if strings.HasPrefix(line, "id:") {
		note.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
	} else if strings.HasPrefix(line, "date:") {
		dateStr := strings.TrimSpace(strings.TrimPrefix(line, "date:"))
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			note.Date = t
		}
	} else if strings.HasPrefix(line, "  - ") {
		// This is a tag or alias item
		tag := strings.TrimSpace(strings.TrimPrefix(line, "  - "))
		note.Tags = append(note.Tags, tag)
	}
}

// parseWorkItem parses a work item line (checkbox format)
func (p *Parser) parseWorkItem(line string) *WorkItem {
	line = strings.TrimSpace(line)

	// Match unchecked: - [ ] task
	if strings.HasPrefix(line, "- [ ] ") {
		return &WorkItem{
			Text:      strings.TrimPrefix(line, "- [ ] "),
			Completed: false,
		}
	}

	// Match checked: - [x] task
	if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ") {
		text := strings.TrimPrefix(line, "- [x] ")
		text = strings.TrimPrefix(text, "- [X] ")
		return &WorkItem{
			Text:      text,
			Completed: true,
		}
	}

	return nil
}

// FindMostRecentNote finds the most recent note before the given date
func (p *Parser) FindMostRecentNote(beforeDate time.Time) (*Note, error) {
	pattern := filepath.Join(p.notesDir, fmt.Sprintf("*-%s.md", p.workplaceName))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	// Parse dates from filenames and sort
	type fileDate struct {
		path string
		date time.Time
	}

	var validFiles []fileDate
	dateRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-.*\.md$`)

	for _, f := range files {
		basename := filepath.Base(f)
		matches := dateRegex.FindStringSubmatch(basename)
		if len(matches) >= 2 {
			if date, err := time.Parse("2006-01-02", matches[1]); err == nil {
				// Only include dates before the target date
				if date.Before(beforeDate) {
					validFiles = append(validFiles, fileDate{path: f, date: date})
				}
			}
		}
	}

	if len(validFiles) == 0 {
		return nil, nil
	}

	// Sort by date descending (most recent first)
	sort.Slice(validFiles, func(i, j int) bool {
		return validFiles[i].date.After(validFiles[j].date)
	})

	// Return the most recent note
	return p.ParseFile(validFiles[0].path)
}

// FindTodayNote finds today's note if it exists
func (p *Parser) FindTodayNote(date time.Time) (*Note, error) {
	filename := GenerateFilename(date, p.workplaceName)
	filePath := filepath.Join(p.notesDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}

	return p.ParseFile(filePath)
}

// NoteExists checks if a note exists for the given date
func (p *Parser) NoteExists(date time.Time) bool {
	filename := GenerateFilename(date, p.workplaceName)
	filePath := filepath.Join(p.notesDir, filename)
	_, err := os.Stat(filePath)
	return err == nil
}
