package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Format int

const (
	FormatTable Format = iota
	FormatJSON
	FormatToon
)

// ansi colors
const (
	red    = "\033[31m"
	green  = "\033[32m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func colorAmount(typ string, amount string) string {
	switch typ {
	case "expense":
		return red + "-" + amount + reset
	case "income", "refund":
		return green + "+" + amount + reset
	default:
		return amount
	}
}

func writeJSON(w io.Writer, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Fprintln(w, string(data))
}

// WriteJSONPublic is an exported version of writeJSON for use by CLI commands.
func WriteJSONPublic(w io.Writer, v interface{}) {
	writeJSON(w, v)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padLeft(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return strings.Repeat(" ", n-len(s)) + s
}
