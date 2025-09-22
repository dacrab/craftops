package view

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Color scheme
var (
	SuccessColor = color.New(color.FgGreen, color.Bold)
	ErrorColor   = color.New(color.FgRed, color.Bold)
	WarningColor = color.New(color.FgYellow, color.Bold)
	InfoColor    = color.New(color.FgCyan, color.Bold)
	HeaderColor  = color.New(color.FgMagenta, color.Bold)
	AccentColor  = color.New(color.FgBlue, color.Bold)
	DimColor     = color.New(color.FgHiBlack)
)

var out io.Writer = os.Stdout

// SetWriter sets the output writer for all view printing
func SetWriter(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	out = w
	color.Output = w
}

// PrintBanner prints a beautiful banner
func PrintBanner(title string) {
	width := 60
	padding := (width - len(title) - 4) / 2

	HeaderColor.Println(strings.Repeat("═", width))
	HeaderColor.Printf("║%s🎮 %s 🎮%s║\n",
		strings.Repeat(" ", padding),
		title,
		strings.Repeat(" ", padding))
	HeaderColor.Println(strings.Repeat("═", width))
	fmt.Fprintln(out)
}

// PrintSection prints a section header
func PrintSection(title string) {
	AccentColor.Printf("\n▶ %s\n", title)
	DimColor.Println(strings.Repeat("─", len(title)+2))
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	SuccessColor.Printf("✅ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	ErrorColor.Printf("❌ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	WarningColor.Printf("⚠️  %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	InfoColor.Printf("ℹ️  %s\n", message)
}

// PrintStep prints a step in a process
func PrintStep(step int, total int, message string) {
	AccentColor.Printf("[%d/%d] ", step, total)
	fmt.Fprintf(out, "%s\n", message)
}

// PrintTable prints a formatted table
func PrintTable(headers []string, rows [][]string) {
	// Calculate column widths
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

	// Print header
	AccentColor.Print("┌")
	for i, width := range widths {
		AccentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			AccentColor.Print("┬")
		}
	}
	AccentColor.Println("┐")

	AccentColor.Print("│")
	for i, header := range headers {
		fmt.Fprintf(out, " %-*s ", widths[i], header)
		AccentColor.Print("│")
	}
	fmt.Fprintln(out)

	AccentColor.Print("├")
	for i, width := range widths {
		AccentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			AccentColor.Print("┼")
		}
	}
	AccentColor.Println("┤")

	// Print rows
	for _, row := range rows {
		AccentColor.Print("│")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(out, " %-*s ", widths[i], cell)
				AccentColor.Print("│")
			}
		}
		fmt.Fprintln(out)
	}

	AccentColor.Print("└")
	for i, width := range widths {
		AccentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			AccentColor.Print("┴")
		}
	}
	AccentColor.Println("┘")
}
