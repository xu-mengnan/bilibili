package file

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// SanitizeSpreadsheetCell neutralizes leading formula markers in user-controlled
// CSV fields. Spreadsheet applications commonly evaluate cells beginning with
// =, +, -, @, tab, or carriage return as formulas/commands.
func SanitizeSpreadsheetCell(value string) string {
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed == "" {
		return value
	}
	switch trimmed[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

func WriteCSV(data [][]string, filePath string) (err error) {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(f)
	for _, row := range data {
		safeRow := make([]string, len(row))
		for i, cell := range row {
			safeRow[i] = SanitizeSpreadsheetCell(cell)
		}
		if err := writer.Write(safeRow); err != nil {
			_ = f.Close()
			return err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = f.Close()
		return fmt.Errorf("flush CSV: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync CSV: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close CSV: %w", err)
	}
	return nil
}

func ReadCSV(filePath string) (records [][]string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(f)
	records, readErr := reader.ReadAll()
	closeErr := f.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close CSV: %w", closeErr)
	}
	return records, nil
}
