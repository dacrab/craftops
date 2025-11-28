package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"

	"craftops/internal/domain"
)

// Terminal handles all terminal output with consistent styling
type Terminal struct {
	out     io.Writer
	errOut  io.Writer
	isTTY   bool
	noColor bool
}

// Colors
var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgCyan, color.Bold)
	headerColor  = color.New(color.FgMagenta, color.Bold)
	accentColor  = color.New(color.FgBlue, color.Bold)
	dimColor     = color.New(color.FgHiBlack)
)

// NewTerminal creates a new terminal output handler
func NewTerminal() *Terminal {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	color.NoColor = !isTTY

	return &Terminal{
		out:     os.Stdout,
		errOut:  os.Stderr,
		isTTY:   isTTY,
		noColor: !isTTY,
	}
}

// NewTerminalWithWriter creates a terminal with custom writers (for testing)
func NewTerminalWithWriter(out, errOut io.Writer, isTTY bool) *Terminal {
	return &Terminal{
		out:     out,
		errOut:  errOut,
		isTTY:   isTTY,
		noColor: !isTTY,
	}
}

// IsTTY returns whether output is a terminal
func (t *Terminal) IsTTY() bool {
	return t.isTTY
}

// Banner prints a decorated banner
func (t *Terminal) Banner(title string) {
	if !t.isTTY {
		fmt.Fprintf(t.out, "%s\n", title)
		return
	}

	width := 60
	padding := (width - len(title) - 4) / 2

	headerColor.Fprintln(t.out, strings.Repeat("═", width))
	headerColor.Fprintf(t.out, "║%s %s %s║\n",
		strings.Repeat(" ", padding),
		title,
		strings.Repeat(" ", padding))
	headerColor.Fprintln(t.out, strings.Repeat("═", width))
	fmt.Fprintln(t.out)
}

// Section prints a section header
func (t *Terminal) Section(title string) {
	if t.isTTY {
		accentColor.Fprintf(t.out, "\n▶ %s\n", title)
		dimColor.Fprintln(t.out, strings.Repeat("─", len(title)+2))
	} else {
		fmt.Fprintf(t.out, "\n== %s ==\n", title)
	}
}

// Success prints a success message
func (t *Terminal) Success(message string) {
	if t.isTTY {
		successColor.Fprintln(t.out, message)
	} else {
		fmt.Fprintf(t.out, "SUCCESS: %s\n", message)
	}
}

// Error prints an error message
func (t *Terminal) Error(message string) {
	if t.isTTY {
		errorColor.Fprintln(t.out, message)
	} else {
		fmt.Fprintf(t.out, "ERROR: %s\n", message)
	}
}

// Warning prints a warning message
func (t *Terminal) Warning(message string) {
	if t.isTTY {
		warningColor.Fprintln(t.out, message)
	} else {
		fmt.Fprintf(t.out, "WARNING: %s\n", message)
	}
}

// Info prints an info message
func (t *Terminal) Info(message string) {
	if t.isTTY {
		infoColor.Fprintln(t.out, message)
	} else {
		fmt.Fprintf(t.out, "INFO: %s\n", message)
	}
}

// Step prints a numbered step
func (t *Terminal) Step(current, total int, message string) {
	if t.isTTY {
		accentColor.Fprintf(t.out, "[%d/%d] ", current, total)
	} else {
		fmt.Fprintf(t.out, "[%d/%d] ", current, total)
	}
	fmt.Fprintln(t.out, message)
}

// Printf prints formatted text
func (t *Terminal) Printf(format string, args ...interface{}) {
	fmt.Fprintf(t.out, format, args...)
}

// Println prints a line
func (t *Terminal) Println(args ...interface{}) {
	fmt.Fprintln(t.out, args...)
}

// AccentSprintf returns accent-colored text
func (t *Terminal) AccentSprintf(format string, args ...interface{}) string {
	if t.isTTY {
		return accentColor.Sprintf(format, args...)
	}
	return fmt.Sprintf(format, args...)
}

// SuccessSprint returns success-colored text
func (t *Terminal) SuccessSprint(text string) string {
	if t.isTTY {
		return successColor.Sprint(text)
	}
	return text
}

// ErrorSprint returns error-colored text
func (t *Terminal) ErrorSprint(text string) string {
	if t.isTTY {
		return errorColor.Sprint(text)
	}
	return text
}

// WarningSprint returns warning-colored text
func (t *Terminal) WarningSprint(text string) string {
	if t.isTTY {
		return warningColor.Sprint(text)
	}
	return text
}

// DimSprint returns dimmed text
func (t *Terminal) DimSprint(text string) string {
	if t.isTTY {
		return dimColor.Sprint(text)
	}
	return text
}

// Table prints a formatted table
func (t *Terminal) Table(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	if t.isTTY {
		t.printTableTTY(headers, rows, widths)
	} else {
		t.printTablePlain(headers, rows, widths)
	}
}

func (t *Terminal) printTableTTY(headers []string, rows [][]string, widths []int) {
	// Top border
	accentColor.Fprint(t.out, "┌")
	for i, width := range widths {
		accentColor.Fprint(t.out, strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Fprint(t.out, "┬")
		}
	}
	accentColor.Fprintln(t.out, "┐")

	// Header
	accentColor.Fprint(t.out, "│")
	for i, header := range headers {
		fmt.Fprintf(t.out, " %-*s ", widths[i], header)
		accentColor.Fprint(t.out, "│")
	}
	fmt.Fprintln(t.out)

	// Header separator
	accentColor.Fprint(t.out, "├")
	for i, width := range widths {
		accentColor.Fprint(t.out, strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Fprint(t.out, "┼")
		}
	}
	accentColor.Fprintln(t.out, "┤")

	// Rows
	for _, row := range rows {
		accentColor.Fprint(t.out, "│")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(t.out, " %-*s ", widths[i], cell)
				accentColor.Fprint(t.out, "│")
			}
		}
		fmt.Fprintln(t.out)
	}

	// Bottom border
	accentColor.Fprint(t.out, "└")
	for i, width := range widths {
		accentColor.Fprint(t.out, strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Fprint(t.out, "┴")
		}
	}
	accentColor.Fprintln(t.out, "┘")
}

func (t *Terminal) printTablePlain(headers []string, rows [][]string, widths []int) {
	// Header
	for i, header := range headers {
		fmt.Fprintf(t.out, "%-*s  ", widths[i], header)
	}
	fmt.Fprintln(t.out)

	// Separator
	for i, width := range widths {
		fmt.Fprint(t.out, strings.Repeat("-", width))
		if i < len(widths)-1 {
			fmt.Fprint(t.out, "  ")
		}
	}
	fmt.Fprintln(t.out)

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(t.out, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(t.out)
	}
}

// HealthCheckTable prints health check results as a table
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
