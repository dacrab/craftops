// Package ui provides styled terminal output for the CLI.
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

// Terminal provides structured output with optional color and formatting.
type Terminal struct {
	out    io.Writer
	errOut io.Writer
	isTTY  bool
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

// NewTerminal creates a terminal linked to stdout/stderr.
func NewTerminal() *Terminal {
	isTTY := term.IsTerminal(int(os.Stdout.Fd())) //nolint:gosec
	color.NoColor = !isTTY
	return &Terminal{out: os.Stdout, errOut: os.Stderr, isTTY: isTTY}
}

// NewTerminalWithWriter creates a terminal with custom writers (for testing).
func NewTerminalWithWriter(out, errOut io.Writer, isTTY bool) *Terminal {
	return &Terminal{out: out, errOut: errOut, isTTY: isTTY}
}

// IsTTY reports whether output is a terminal.
func (t *Terminal) IsTTY() bool { return t.isTTY }

// Banner prints a prominent header.
func (t *Terminal) Banner(title string) {
	if !t.isTTY {
		_, _ = fmt.Fprintf(t.out, "%s\n", title)
		return
	}
	width := 60
	padding := (width - len(title) - 4) / 2
	_, _ = headerColor.Fprintln(t.out, strings.Repeat("═", width))
	_, _ = headerColor.Fprintf(t.out, "║%s %s %s║\n",
		strings.Repeat(" ", padding), title, strings.Repeat(" ", padding))
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
func (t *Terminal) Successf(format string, args ...interface{}) {
	t.Success(fmt.Sprintf(format, args...))
}

// Error prints an error message.
func (t *Terminal) Error(message string) { t.printMsg(errorColor, "ERROR", message) }

// Errorf prints a formatted error message.
func (t *Terminal) Errorf(format string, args ...interface{}) {
	t.Error(fmt.Sprintf(format, args...))
}

// Warning prints a warning message.
func (t *Terminal) Warning(message string) { t.printMsg(warningColor, "WARNING", message) }

// Warningf prints a formatted warning message.
func (t *Terminal) Warningf(format string, args ...interface{}) {
	t.Warning(fmt.Sprintf(format, args...))
}

// Info prints an info message.
func (t *Terminal) Info(message string) { t.printMsg(infoColor, "INFO", message) }

// Infof prints a formatted info message.
func (t *Terminal) Infof(format string, args ...interface{}) {
	t.Info(fmt.Sprintf(format, args...))
}

func (t *Terminal) printMsg(c *color.Color, label, msg string) {
	if t.isTTY {
		_, _ = c.Fprintln(t.out, msg)
	} else {
		_, _ = fmt.Fprintf(t.out, "%s: %s\n", label, msg)
	}
}

// Step prints a progress indicator like [1/5].
func (t *Terminal) Step(current, total int, message string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "[%d/%d] ", current, total)
	} else {
		_, _ = fmt.Fprintf(t.out, "[%d/%d] ", current, total)
	}
	_, _ = fmt.Fprintln(t.out, message)
}

// Printf writes formatted output.
func (t *Terminal) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t.out, format, args...)
}

// Println writes a line of output.
func (t *Terminal) Println(args ...interface{}) {
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

// Table renders a formatted table.
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
	table.Header(stringsToAny(headers)...)
	for _, row := range rows {
		if err := table.Append(stringsToAny(row)...); err != nil {
			_, _ = fmt.Fprintf(t.errOut, "Table append error: %v\n", err)
		}
	}
	if err := table.Render(); err != nil {
		_, _ = fmt.Fprintf(t.errOut, "Table render error: %v\n", err)
	}
}

func stringsToAny(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i := range strs {
		result[i] = strs[i]
	}
	return result
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
