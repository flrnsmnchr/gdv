package diff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DiffLine struct {
	Gutter string
	Text   string
	OldNo  int
	NewNo  int
}

var hunkHeaderRE = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func NumberedDiffLines(diff string) []DiffLine {
	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")
	oldNo := 0
	newNo := 0
	maxNo := 0
	for _, line := range lines {
		if oldNo > maxNo {
			maxNo = oldNo
		}
		if newNo > maxNo {
			maxNo = newNo
		}
		switch {
		case strings.HasPrefix(line, "@@"):
			continue
		case strings.HasPrefix(line, "diff --git"), strings.HasPrefix(line, "index "), strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
			continue
		case len(line) == 0:
			continue
		case line[0] == ' ':
			oldNo++
			newNo++
		case line[0] == '-':
			oldNo++
		case line[0] == '+':
			newNo++
		}
	}

	width := numberWidth(maxNo)
	oldNo = 0
	newNo = 0
	out := make([]DiffLine, 0, len(lines))
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"), strings.HasPrefix(line, "index "), strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
			out = append(out, DiffLine{Gutter: blankDiffGutter(width), Text: line})
		case strings.HasPrefix(line, "@@"):
			if oldStart, newStart, ok := parseHunkHeader(line); ok {
				oldNo = oldStart
				newNo = newStart
			}
			out = append(out, DiffLine{Gutter: blankDiffGutter(width), Text: line})
		case len(line) == 0:
			out = append(out, DiffLine{Gutter: blankDiffGutter(width), Text: line})
		case line[0] == ' ':
			out = append(out, DiffLine{
				Gutter: fmt.Sprintf("%*d | %*d", width, oldNo, width, newNo),
				Text:   line,
				OldNo:  oldNo,
				NewNo:  newNo,
			})
			oldNo++
			newNo++
		case line[0] == '-':
			out = append(out, DiffLine{
				Gutter: fmt.Sprintf("%*d | %*s", width, oldNo, width, ""),
				Text:   line,
				OldNo:  oldNo,
			})
			oldNo++
		case line[0] == '+':
			out = append(out, DiffLine{
				Gutter: fmt.Sprintf("%*s | %*d", width, "", width, newNo),
				Text:   line,
				NewNo:  newNo,
			})
			newNo++
		default:
			out = append(out, DiffLine{Gutter: blankDiffGutter(width), Text: line})
		}
	}
	return out
}

func parseHunkHeader(line string) (int, int, bool) {
	match := hunkHeaderRE.FindStringSubmatch(line)
	if len(match) != 3 {
		return 0, 0, false
	}
	oldNo, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, false
	}
	newNo, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, 0, false
	}
	return oldNo, newNo, true
}

func DiffHunkOffsets(diff string) []int {
	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")
	offsets := make([]int, 0, 4)
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		if (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) &&
			!strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "---") {
			if i == 0 {
				offsets = append(offsets, i)
			} else {
				prev := lines[i-1]
				if !(strings.HasPrefix(prev, "+") || strings.HasPrefix(prev, "-")) {
					offsets = append(offsets, i)
				}
			}
		}
	}
	return offsets
}

func SideSource(diff string, marker byte) string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") {
			continue
		}
		if strings.HasPrefix(line, "@@") {
			continue
		}
		if line[0] == ' ' {
			out = append(out, line[1:])
			continue
		}
		if line[0] == marker {
			out = append(out, line[1:])
		}
	}
	if len(out) == 0 {
		return "(empty)"
	}
	return strings.Join(out, "\n")
}

func OldSource(diff string) string {
	return SideSource(diff, '-')
}

func NewSource(diff string) string {
	return SideSource(diff, '+')
}

func numberWidth(maxNo int) int {
	if maxNo < 1 {
		return 1
	}
	width := 0
	for maxNo > 0 {
		width++
		maxNo /= 10
	}
	if width < 4 {
		return 4
	}
	return width
}

func blankDiffGutter(width int) string {
	return fmt.Sprintf("%*s | %*s", width, "", width, "")
}
