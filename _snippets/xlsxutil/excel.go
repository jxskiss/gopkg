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

var trySheetNameKeys = []string{"SheetName", "sheetName", "sheet_name", "SheetKey", "sheetKey", "sheet_key", "Sheet", "sheet"}

func SaveExcelFile[T ~map[string]any](filename string, data []T, columns []string) error {
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
	var sheetKey string
	for _, x := range trySheetNameKeys {
		if _, ok := data[0][x]; ok {
			sheetKey = x
			break
		}
	}

	// Split and save sheets.
	var sheetNames = []any{defaultSheetName}
	var sheetRecords map[any][]T
	if sheetKey == "" {
		sheetRecords = map[any][]T{
			defaultSheetName: data,
		}
	} else {
		sheetNames, sheetRecords = GroupMapRecords(data, sheetKey)
	}
	for _, sheetNameAny := range sheetNames {
		sheetName := fmt.Sprint(sheetNameAny)
		records := sheetRecords[sheetNameAny]
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

func WriteExcelSheet[T ~map[string]any](file *excelize.File, sheetName string, records []T, columns []string) error {
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

func GroupMapRecords[T ~map[string]any](records []T, col string) (groupKeys []any, groupRecords map[any][]T) {
	groupRecords = make(map[any][]T)
	for _, record := range records {
		key := record[col]
		if key == nil || key == "" {
			key = defaultSheetName
		}
		if !isInSlice(groupKeys, key) {
			groupKeys = append(groupKeys, key)
		}
		groupRecords[key] = append(groupRecords[key], record)
	}
	return groupKeys, groupRecords
}

func isInSlice(s []any, x any) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}
