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
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
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

func (t *Terminal) IsTTY() bool { return t.isTTY }

// Banner prints a prominent centered header with double-line borders
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

// Section prints a secondary header with an arrow indicator
func (t *Terminal) Section(title string) {
	if t.isTTY {
		accentColor.Fprintf(t.out, "\n▶ %s\n", title)
		dimColor.Fprintln(t.out, strings.Repeat("─", len(title)+2))
	} else {
		fmt.Fprintf(t.out, "\n== %s ==\n", title)
	}
}

func (t *Terminal) Success(message string) { t.printMsg(successColor, "SUCCESS", message) }
func (t *Terminal) Error(message string)   { t.printMsg(errorColor, "ERROR", message) }
func (t *Terminal) Warning(message string) { t.printMsg(warningColor, "WARNING", message) }
func (t *Terminal) Info(message string)    { t.printMsg(infoColor, "INFO", message) }

func (t *Terminal) printMsg(c *color.Color, label, msg string) {
	if t.isTTY {
		c.Fprintln(t.out, msg)
	} else {
		fmt.Fprintf(t.out, "%s: %s\n", label, msg)
	}
}

// Step prints a progress indicator like [1/5]
func (t *Terminal) Step(current, total int, message string) {
	if t.isTTY {
		accentColor.Fprintf(t.out, "[%d/%d] ", current, total)
	} else {
		fmt.Fprintf(t.out, "[%d/%d] ", current, total)
	}
	fmt.Fprintln(t.out, message)
}

func (t *Terminal) Printf(format string, args ...interface{}) { fmt.Fprintf(t.out, format, args...) }
func (t *Terminal) Println(args ...interface{})               { fmt.Fprintln(t.out, args...) }

func (t *Terminal) AccentSprintf(format string, args ...interface{}) string {
	if t.isTTY {
		return accentColor.Sprintf(format, args...)
	}
	return fmt.Sprintf(format, args...)
}

func (t *Terminal) SuccessSprint(text string) string {
	if t.isTTY {
		return successColor.Sprint(text)
	}
	return text
}

func (t *Terminal) ErrorSprint(text string) string {
	if t.isTTY {
		return errorColor.Sprint(text)
	}
	return text
}

func (t *Terminal) WarningSprint(text string) string {
	if t.isTTY {
		return warningColor.Sprint(text)
	}
	return text
}

func (t *Terminal) DimSprint(text string) string {
	if t.isTTY {
		return dimColor.Sprint(text)
	}
	return text
}

// Table prints a dynamically-sized table based on terminal capabilities
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
	line := func(left, mid, right, fill string) {
		if t.isTTY {
			accentColor.Fprint(t.out, left)
		} else {
			fmt.Fprint(t.out, left)
		}
		for i, w := range widths {
			fmt.Fprint(t.out, strings.Repeat(fill, w+2))
			if i < len(widths)-1 {
				if t.isTTY {
					accentColor.Fprint(t.out, mid)
				} else {
					fmt.Fprint(t.out, mid)
				}
			}
		}
		if t.isTTY {
			accentColor.Fprintln(t.out, right)
		} else {
			fmt.Fprintln(t.out, right)
		}
	}

	line("┌", "┬", "┐", "─")
	if t.isTTY {
		accentColor.Fprint(t.out, "│")
	} else {
		fmt.Fprint(t.out, "│")
	}
	for i, h := range headers {
		fmt.Fprintf(t.out, " %-*s ", widths[i], h)
		if t.isTTY {
			accentColor.Fprint(t.out, "│")
		} else {
			fmt.Fprint(t.out, "│")
		}
	}
	fmt.Fprintln(t.out)
	line("├", "┼", "┤", "─")

	for _, row := range rows {
		if t.isTTY {
			accentColor.Fprint(t.out, "│")
		} else {
			fmt.Fprint(t.out, "│")
		}
		for i, c := range row {
			if i < len(widths) {
				fmt.Fprintf(t.out, " %-*s ", widths[i], c)
				if t.isTTY {
					accentColor.Fprint(t.out, "│")
				} else {
					fmt.Fprint(t.out, "│")
				}
			}
		}
		fmt.Fprintln(t.out)
	}
	line("└", "┴", "┘", "─")
}

func (t *Terminal) printTablePlain(headers []string, rows [][]string, widths []int) {
	for i, h := range headers {
		fmt.Fprintf(t.out, "%-*s  ", widths[i], h)
	}
	fmt.Fprintln(t.out)
	for i, w := range widths {
		fmt.Fprint(t.out, strings.Repeat("-", w))
		if i < len(widths)-1 {
			fmt.Fprint(t.out, "  ")
		}
	}
	fmt.Fprintln(t.out)
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(t.out, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(t.out)
	}
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
