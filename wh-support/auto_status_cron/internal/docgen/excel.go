package docgen

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ExcelSheetData struct {
	Name    string
	Columns []ExcelColumnData
}

type ExcelColumnData struct {
	Name   string
	Values []any
}

func (s *Service) GenerateExcel(sheets ...ExcelSheetData) ([]byte, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			logrus.Errorf("error closing file: %v", err)
		}
	}()

	for i, sheet := range sheets {
		if i == 0 {
			if err := f.SetSheetName("Sheet1", sheet.Name); err != nil {
				return nil, fmt.Errorf("can't set sheet name: %w", err)
			}
		} else {
			if _, err := f.NewSheet(sheet.Name); err != nil {
				return nil, fmt.Errorf("can't create sheet %s: %w", sheet.Name, err)
			}
		}

		if err := writeSheetData(f, sheet); err != nil {
			return nil, fmt.Errorf("can't write sheet %s: %w", sheet.Name, err)
		}
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, fmt.Errorf("can't write to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

func writeSheetData(f *excelize.File, sheet ExcelSheetData) error {
	for i, col := range sheet.Columns {
		colName := fmt.Sprintf("%c", 'A'+i)

		if err := f.SetCellValue(sheet.Name, fmt.Sprintf("%s1", colName), col.Name); err != nil {
			return fmt.Errorf("can't set column name: %w", err)
		}

		if err := f.SetSheetCol(sheet.Name, fmt.Sprintf("%s2", colName), &col.Values); err != nil {
			return fmt.Errorf("can't set column's values: %w", err)
		}
	}
	return nil
}
