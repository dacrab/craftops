// Package ui provides terminal output formatting and styling
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"golang.org/x/term"

	"craftops/internal/domain"
)

// Terminal provides structured and styled output to the console
type Terminal struct {
	out     io.Writer
	errOut  io.Writer
	isTTY   bool
	noColor bool
}

// Global color definitions for consistent UI branding
var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgCyan, color.Bold)
	headerColor  = color.New(color.FgMagenta, color.Bold)
	accentColor  = color.New(color.FgBlue, color.Bold)
	dimColor     = color.New(color.FgHiBlack)
)

// NewTerminal initializes a terminal linked to standard output
func NewTerminal() *Terminal {
	isTTY := term.IsTerminal(int(os.Stdout.Fd())) //nolint:gosec // uintptr→int safe: file descriptors fit in int on all supported platforms
	color.NoColor = !isTTY

	return &Terminal{
		out:     os.Stdout,
		errOut:  os.Stderr,
		isTTY:   isTTY,
		noColor: !isTTY,
	}
}

// NewTerminalWithWriter allows injecting custom writers for testing or redirection
func NewTerminalWithWriter(out, errOut io.Writer, isTTY bool) *Terminal {
	return &Terminal{
		out:     out,
		errOut:  errOut,
		isTTY:   isTTY,
		noColor: !isTTY,
	}
}

// IsTTY returns whether the terminal is a TTY
func (t *Terminal) IsTTY() bool { return t.isTTY }

// Banner prints a prominent centered header with double-line borders
func (t *Terminal) Banner(title string) {
	if !t.isTTY {
		_, _ = fmt.Fprintf(t.out, "%s\n", title) //nolint:errcheck // stdout errors are extremely rare
		return
	}

	width := 60
	padding := (width - len(title) - 4) / 2

	_, _ = headerColor.Fprintln(t.out, strings.Repeat("═", width)) //nolint:errcheck
	_, _ = headerColor.Fprintf(t.out, "║%s %s %s║\n", //nolint:errcheck
		strings.Repeat(" ", padding),
		title,
		strings.Repeat(" ", padding))
	_, _ = headerColor.Fprintln(t.out, strings.Repeat("═", width)) //nolint:errcheck
	_, _ = fmt.Fprintln(t.out) //nolint:errcheck
}

// Section prints a secondary header with an arrow indicator
func (t *Terminal) Section(title string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "\n▶ %s\n", title) //nolint:errcheck
		_, _ = dimColor.Fprintln(t.out, strings.Repeat("─", len(title)+2)) //nolint:errcheck
	} else {
		_, _ = fmt.Fprintf(t.out, "\n== %s ==\n", title) //nolint:errcheck
	}
}

// Success prints a success message
func (t *Terminal) Success(message string) { t.printMsg(successColor, "SUCCESS", message) }

// Error prints an error message
func (t *Terminal) Error(message string) { t.printMsg(errorColor, "ERROR", message) }

// Warning prints a warning message
func (t *Terminal) Warning(message string) { t.printMsg(warningColor, "WARNING", message) }

// Info prints an info message
func (t *Terminal) Info(message string) { t.printMsg(infoColor, "INFO", message) }

func (t *Terminal) printMsg(c *color.Color, label, msg string) {
	if t.isTTY {
		_, _ = c.Fprintln(t.out, msg) //nolint:errcheck
	} else {
		_, _ = fmt.Fprintf(t.out, "%s: %s\n", label, msg) //nolint:errcheck
	}
}

// Step prints a progress indicator like [1/5]
func (t *Terminal) Step(current, total int, message string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "[%d/%d] ", current, total) //nolint:errcheck
	} else {
		_, _ = fmt.Fprintf(t.out, "[%d/%d] ", current, total) //nolint:errcheck
	}
	_, _ = fmt.Fprintln(t.out, message) //nolint:errcheck
}

// Printf formats and prints to the terminal output
func (t *Terminal) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t.out, format, args...) //nolint:errcheck
}

// Println prints arguments to the terminal output
func (t *Terminal) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(t.out, args...) //nolint:errcheck
}

// AccentSprintf formats a string with accent color
func (t *Terminal) AccentSprintf(format string, args ...interface{}) string {
	if t.isTTY {
		return accentColor.Sprintf(format, args...)
	}
	return fmt.Sprintf(format, args...)
}

// SuccessSprint returns text formatted with success color
func (t *Terminal) SuccessSprint(text string) string {
	return t.sprintWithColor(text, successColor)
}

// ErrorSprint returns text formatted with error color
func (t *Terminal) ErrorSprint(text string) string {
	return t.sprintWithColor(text, errorColor)
}

// WarningSprint returns text formatted with warning color
func (t *Terminal) WarningSprint(text string) string {
	return t.sprintWithColor(text, warningColor)
}

// DimSprint returns text formatted with dim color
func (t *Terminal) DimSprint(text string) string {
	return t.sprintWithColor(text, dimColor)
}

// sprintWithColor applies color formatting if TTY is enabled
func (t *Terminal) sprintWithColor(text string, c *color.Color) string {
	if t.isTTY {
		return c.Sprint(text)
	}
	return text
}

// Table prints a dynamically-sized table using github.com/olekukonko/tablewriter
func (t *Terminal) Table(headers []string, rows [][]string) {
	var opts []tablewriter.Option
	if t.isTTY {
		opts = []tablewriter.Option{
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.Border{Left: tw.On, Top: tw.On, Right: tw.On, Bottom: tw.On},
			}),
			tablewriter.WithHeaderAlignment(tw.AlignCenter),
			tablewriter.WithHeaderAutoFormat(tw.On),
		}
	} else {
		opts = []tablewriter.Option{
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Top: tw.Off, Right: tw.Off, Bottom: tw.Off},
			}),
		}
	}

	table := tablewriter.NewTable(t.out, opts...)
	table.Header(convertToInterfaceSlice(headers)...)
	for _, row := range rows {
		if err := table.Append(convertToInterfaceSlice(row)...); err != nil {
			// Log error but continue - best effort rendering
			_, _ = fmt.Fprintf(t.errOut, "Table append error: %v\n", err) //nolint:errcheck
		}
	}
	if err := table.Render(); err != nil {
		_, _ = fmt.Fprintf(t.errOut, "Table render error: %v\n", err) //nolint:errcheck
	}
}

// convertToInterfaceSlice converts []string to []interface{} for tablewriter
func convertToInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i := range strs {
		result[i] = strs[i]
	}
	return result
}

// HealthCheckTable outputs a specialized table for diagnostic results
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
