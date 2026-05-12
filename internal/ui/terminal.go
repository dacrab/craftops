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

func NewTerminal() *Terminal {
	isTTY := term.IsTerminal(int(os.Stdout.Fd())) //nolint:gosec
	color.NoColor = !isTTY
	return &Terminal{out: os.Stdout, errOut: os.Stderr, isTTY: isTTY}
}

func NewTerminalWithWriter(out, errOut io.Writer, isTTY bool) *Terminal {
	return &Terminal{out: out, errOut: errOut, isTTY: isTTY}
}

func (t *Terminal) IsTTY() bool { return t.isTTY }

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

func (t *Terminal) Section(title string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "\n▶ %s\n", title)
		_, _ = dimColor.Fprintln(t.out, strings.Repeat("─", len(title)+2))
	} else {
		_, _ = fmt.Fprintf(t.out, "\n== %s ==\n", title)
	}
}

func (t *Terminal) Success(message string) { t.printMsg(successColor, "SUCCESS", message) }
func (t *Terminal) Successf(format string, args ...interface{}) {
	t.Success(fmt.Sprintf(format, args...))
}

func (t *Terminal) Error(message string) { t.printMsg(errorColor, "ERROR", message) }
func (t *Terminal) Errorf(format string, args ...interface{}) {
	t.Error(fmt.Sprintf(format, args...))
}

func (t *Terminal) Warning(message string) { t.printMsg(warningColor, "WARNING", message) }
func (t *Terminal) Warningf(format string, args ...interface{}) {
	t.Warning(fmt.Sprintf(format, args...))
}

func (t *Terminal) Info(message string) { t.printMsg(infoColor, "INFO", message) }
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

func (t *Terminal) Step(current, total int, message string) {
	if t.isTTY {
		_, _ = accentColor.Fprintf(t.out, "[%d/%d] ", current, total)
	} else {
		_, _ = fmt.Fprintf(t.out, "[%d/%d] ", current, total)
	}
	_, _ = fmt.Fprintln(t.out, message)
}

func (t *Terminal) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(t.out, format, args...)
}

func (t *Terminal) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(t.out, args...)
}

func (t *Terminal) SuccessSprint(text string) string { return t.sprintWithColor(text, successColor) }
func (t *Terminal) ErrorSprint(text string) string   { return t.sprintWithColor(text, errorColor) }
func (t *Terminal) WarningSprint(text string) string  { return t.sprintWithColor(text, warningColor) }
func (t *Terminal) DimSprint(text string) string      { return t.sprintWithColor(text, dimColor) }

func (t *Terminal) sprintWithColor(text string, c *color.Color) string {
	if t.isTTY {
		return c.Sprint(text)
	}
	return text
}

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

// stringsToAny converts []string to []interface{} as required by the tablewriter API.
func stringsToAny(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i := range strs {
		result[i] = strs[i]
	}
	return result
}

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
