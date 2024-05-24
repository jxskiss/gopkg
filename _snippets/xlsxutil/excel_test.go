package xlsxutil

import (
	"testing"
)

func TestSaveExcelFile(t *testing.T) {
	records := []map[string]any{
		{
			"A":     123,
			"B":     "abc",
			"C":     &Link{Value: "a link", URL: "https://www.example.com/123"},
			"Sheet": "SheetTest",
		},
		{
			"A": "def",
			"B": 456,
			"C": Link{Value: "another link", URL: "https://www.example.com/456"},
		},
	}
	err := SaveExcelFile("./testout/test_save_excel_file.xlsx", records, []string{
		"A", "B", "C",
	})
	if err != nil {
		t.Errorf("failed to save Excel file: %v", err)
	}
}
