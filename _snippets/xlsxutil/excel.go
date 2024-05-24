package xlsxutil

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"

	"github.com/jxskiss/gopkg/v2/easy"
)

const defaultSheetName = "Sheet1"

type Link struct {
	Value any    `json:"value" yaml:"value"`
	URL   string `json:"url" yaml:"url"`
}

func SaveExcelFile(filename string, data []map[string]any, columns []string) error {
	dirPath := filepath.Dir(filename)
	if dirPath != "/" && dirPath != "." {
		err := easy.CreateNonExistingFolder(dirPath, 0o755)
		if err != nil {
			return fmt.Errorf("create folder: %w", err)
		}
	}

	file := excelize.NewFile()
	defer file.Close()

	// Detect sheet name column.
	sheetKey := "Sheet"
	if _, ok := data[0][sheetKey]; !ok {
		sheetKey = "sheet"
		if _, ok = data[0][sheetKey]; !ok {
			sheetKey = ""
		}
	}

	// Split and save sheets.
	var sheetRecords map[any][]map[string]any
	if sheetKey == "" {
		sheetRecords = map[any][]map[string]any{
			defaultSheetName: data,
		}
	} else {
		sheetRecords = GroupMapRecords(data, sheetKey)
	}
	for sheetNameAny, records := range sheetRecords {
		sheetName := fmt.Sprint(sheetNameAny)
		err := WriteExcelSheet(file, sheetName, records, columns)
		if err != nil {
			return fmt.Errorf("write sheet: %w", err)
		}
	}

	err := file.SaveAs(filename)
	if err != nil {
		return fmt.Errorf("save Excel file: %w", err)
	}
	return nil
}

func WriteExcelSheet(file *excelize.File, sheetName string, records []map[string]any, columns []string) error {
	if sheetName == "" {
		sheetName = defaultSheetName
	}
	sheetIdx, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return fmt.Errorf("sheetName is invalid: %w", err)
	}
	if sheetIdx == -1 {
		sheetIdx, err = file.NewSheet(sheetName)
		if err != nil {
			return fmt.Errorf("create new sheet: %w", err)
		}
	}
	file.SetActiveSheet(sheetIdx)

	// Write the column headers.
	err = file.SetSheetRow(sheetName, "A1", &columns)
	if err != nil {
		return fmt.Errorf("write column headers: %w", err)
	}

	// Write the data rows.
	for i, record := range records {
		rowIdx := i + 2
		colIdx := 1
		values, hyperLinks := convExcelValues(record, columns, rowIdx, colIdx)
		cell := mustToCellName(colIdx, rowIdx)
		err = file.SetSheetRow(sheetName, cell, &values)
		if err != nil {
			return fmt.Errorf("write record at index %d to sheet: %w", i, err)
		}
		if len(hyperLinks) > 0 {
			for cell, link := range hyperLinks {
				err = file.SetCellHyperLink(sheetName, cell, link, "External")
				if err != nil {
					return fmt.Errorf("set cell %s hyper link: %w", cell, err)
				}
			}
		}
	}
	return nil
}

func mustToCellName(col, row int, abs ...bool) string {
	cell, err := excelize.CoordinatesToCellName(col, row, abs...)
	if err != nil {
		panic(fmt.Sprintf("cannot convert coordinates to cellName: %v", err))
	}
	return cell
}

func convExcelValues(m map[string]any, columns []string, rowIdx, colIdx int) (
	values []any, hyperLinks map[string]string) {
	for i, col := range columns {
		var linkURL string
		val := m[col]
		if link, ok := val.(Link); ok {
			val = link.Value
			linkURL = link.URL
		} else if lp, ok := val.(*Link); ok {
			val = lp.Value
			linkURL = lp.URL
		}
		values = append(values, val)
		if linkURL != "" {
			if hyperLinks == nil {
				hyperLinks = make(map[string]string)
			}
			cell := mustToCellName(colIdx+i, rowIdx)
			hyperLinks[cell] = linkURL
		}
	}
	return values, hyperLinks
}

func GroupMapRecords(records []map[string]any, col string) map[any][]map[string]any {
	out := make(map[any][]map[string]any)
	for _, record := range records {
		key := record[col]
		out[key] = append(out[key], record)
	}
	return out
}
