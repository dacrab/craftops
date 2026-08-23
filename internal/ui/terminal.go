// Package ui provides styled terminal output for the CLI.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"golang.org/x/term"

	"craftops/internal/domain"
)

// Terminal provides structured output with optional color and formatting.
type Terminal struct {
	out   io.Writer
	isTTY bool
}

var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgCyan, color.Bold)
	headerColor  = color.New(color.FgMagenta, color.Bold)
	accentColor  = color.New(color.FgBlue, color.Bold)
	dimColor     = color.New(color.FgHiBlack)
)

// NewTerminal creates a terminal linked to stdout.
func NewTerminal() *Terminal {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	return &Terminal{out: os.Stdout, isTTY: isTTY}
}

// NewTerminalWithWriter creates a terminal with a custom writer (for testing).
func NewTerminalWithWriter(out io.Writer, isTTY bool) *Terminal {
	return &Terminal{out: out, isTTY: isTTY}
}

// TimeFormat is the human-readable timestamp layout used across the CLI.
const TimeFormat = "2006-01-02 15:04:05"

// FormatSize returns a human-readable file size (e.g. "4.2 MB").
func FormatSize(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

// CheckPath verifies if a path exists and is a directory.
func CheckPath(name, path string) domain.HealthCheck {
	info, err := os.Stat(path)
	if err != nil {
		return domain.HealthCheck{Name: name, Status: domain.StatusWarn, Message: "Does not exist"}
	}
	if !info.IsDir() {
		return domain.HealthCheck{Name: name, Status: domain.StatusError, Message: "Not a directory"}
	}
	return domain.HealthCheck{Name: name, Status: domain.StatusOK, Message: "OK"}
}

// Banner prints a prominent header.
func (t *Terminal) Banner(title string) {
	if !t.isTTY {
		_, _ = fmt.Fprintf(t.out, "%s\n", title)
		return
	}
	width := 60
	gap := width - utf8.RuneCountInString(title) - 4
	if gap < 0 {
		gap = 0
	}
	left := gap / 2
	right := gap - left
	_, _ = headerColor.Fprintln(t.out, strings.Repeat("═", width))
	_, _ = headerColor.Fprintf(t.out, "║%s %s %s║\n",
		strings.Repeat(" ", left), title, strings.Repeat(" ", right))
	_, _ = headerColor.Fprintln(t.out, strings.Repeat("═", width))
	_, _ = fmt.Fprintln(t.out)
}

// Section prints a secondary header.
func (t *Terminal) Section(title string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "\n▶ %s\n", title)
		_, _ = dimColor.Fprintln(t.out, strings.Repeat("─", len(title)+2))
	} else {
		_, _ = fmt.Fprintf(t.out, "\n== %s ==\n", title)
	}
}

// Success prints a success message.
func (t *Terminal) Success(message string) { t.printMsg(successColor, "SUCCESS", message) }

// Successf prints a formatted success message.
func (t *Terminal) Successf(format string, args ...any) {
	t.Success(fmt.Sprintf(format, args...))
}

// Error prints an error message.
func (t *Terminal) Error(message string) { t.printMsg(errorColor, "ERROR", message) }

// Errorf prints a formatted error message.
func (t *Terminal) Errorf(format string, args ...any) {
	t.Error(fmt.Sprintf(format, args...))
}

// Warning prints a warning message.
func (t *Terminal) Warning(message string) { t.printMsg(warningColor, "WARNING", message) }

// Warningf prints a formatted warning message.
func (t *Terminal) Warningf(format string, args ...any) {
	t.Warning(fmt.Sprintf(format, args...))
}

// Info prints an info message.
func (t *Terminal) Info(message string) { t.printMsg(infoColor, "INFO", message) }

func (t *Terminal) printMsg(c *color.Color, label, msg string) {
	if t.isTTY {
		_, _ = c.Fprintln(t.out, msg)
	} else {
		_, _ = fmt.Fprintf(t.out, "%s: %s\n", label, msg)
	}
}

// Printf writes formatted output.
func (t *Terminal) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(t.out, format, args...)
}

// Println writes a line of output.
func (t *Terminal) Println(args ...any) {
	_, _ = fmt.Fprintln(t.out, args...)
}

// SuccessSprint returns text with success color applied.
func (t *Terminal) SuccessSprint(text string) string { return t.sprintWithColor(text, successColor) }

// ErrorSprint returns text with error color applied.
func (t *Terminal) ErrorSprint(text string) string { return t.sprintWithColor(text, errorColor) }

// WarningSprint returns text with warning color applied.
func (t *Terminal) WarningSprint(text string) string { return t.sprintWithColor(text, warningColor) }

// DimSprint returns text with dim color applied.
func (t *Terminal) DimSprint(text string) string { return t.sprintWithColor(text, dimColor) }

func (t *Terminal) sprintWithColor(text string, c *color.Color) string {
	if t.isTTY {
		return c.Sprint(text)
	}
	return text
}

// Table renders a simple aligned table.
func (t *Terminal) Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = visibleLen(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && visibleLen(cell) > widths[i] {
				widths[i] = visibleLen(cell)
			}
		}
	}

	formatRow := func(cells []string) string {
		var b strings.Builder
		for i, cell := range cells {
			if i > 0 {
				b.WriteString("  ")
			}
			if i < len(widths) {
				b.WriteString(pad(cell, widths[i]))
			} else {
				b.WriteString(cell)
			}
		}
		return strings.TrimRight(b.String(), " ")
	}

	_, _ = fmt.Fprintln(t.out, formatRow(headers))
	if t.isTTY {
		_, _ = fmt.Fprintln(t.out, strings.Repeat("─", len(formatRow(headers))))
	}
	for _, row := range rows {
		_, _ = fmt.Fprintln(t.out, formatRow(row))
	}
}

// visibleLen returns the string length ignoring ANSI escape sequences.
func visibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		n++
		i++
	}
	return n
}

// pad pads s to width w using visible (non-ANSI) length.
func pad(s string, w int) string {
	if v := visibleLen(s); v < w {
		return s + strings.Repeat(" ", w-v)
	}
	return s
}

// HealthCheckTable renders a diagnostic results table with colored status.
func (t *Terminal) HealthCheckTable(checks []domain.HealthCheck) {
	headers := []string{"Component", "Status", "Details"}
	rows := make([][]string, len(checks))
	for i, check := range checks {
		status := string(check.Status)
		switch check.Status {
		case domain.StatusOK:
			status = t.SuccessSprint(status)
		case domain.StatusWarn:
			status = t.WarningSprint(status)
		case domain.StatusError:
			status = t.ErrorSprint(status)
		}
		rows[i] = []string{check.Name, status, check.Message}
	}
	t.Table(headers, rows)
}
