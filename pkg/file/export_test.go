package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeSpreadsheetCell(t *testing.T) {
	tests := map[string]string{
		"=1+1":      "'=1+1",
		"+SUM(A1)":  "'+SUM(A1)",
		"-10+20":    "'-10+20",
		"@cmd":      "'@cmd",
		"  =1+1":    "'  =1+1",
		"plain text": "plain text",
		"123":        "123",
	}

	for input, want := range tests {
		if got := SanitizeSpreadsheetCell(input); got != want {
			t.Fatalf("SanitizeSpreadsheetCell(%q)=%q want=%q", input, got, want)
		}
	}
}

func TestWriteCSVNeutralizesFormulaCells(t *testing.T) {
	path := filepath.Join(t.TempDir(), "comments.csv")
	data := [][]string{
		{"user", "comment"},
		{"=HYPERLINK("https://example.com")", "+1+1"},
	}
	if err := WriteCSV(data, path); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "'=HYPERLINK") || !strings.Contains(text, "'+1+1") {
		t.Fatalf("CSV formula cells were not neutralized: %s", text)
	}

	rows, err := ReadCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	if rows[1][0][0] != '\'' || rows[1][1][0] != '\'' {
		t.Fatalf("sanitized CSV values not preserved: %#v", rows[1])
	}
}

func TestExcelRoundTripTreatsFormulaLikeTextAsString(t *testing.T) {
	path := filepath.Join(t.TempDir(), "comments.xlsx")
	input := [][]string{{"=1+1", "@command"}}
	if err := WriteExcel(input, path); err != nil {
		t.Fatal(err)
	}

	rows, err := ReadExcel(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0]) != 2 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
	if rows[0][0] != "=1+1" || rows[0][1] != "@command" {
		t.Fatalf("formula-like strings changed: %#v", rows[0])
	}
}
