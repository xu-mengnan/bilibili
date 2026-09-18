package file

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func WriteExcel(data [][]string, filePath string) (err error) {
	f := excelize.NewFile()
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close Excel file: %w", closeErr)
		}
	}()

	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return err
	}

	for i, row := range data {
		for j, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				return err
			}
			// SetCellValue with a string stores a string cell; formulas are only
			// created through SetCellFormula, so comment text remains literal.
			if err := f.SetCellValue("Sheet1", cellName, cell); err != nil {
				return err
			}
		}
	}

	f.SetActiveSheet(index)
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}

func ReadExcel(filePath string) (rows [][]string, err error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close Excel file: %w", closeErr)
		}
	}()

	rows, err = f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	return rows, nil
}
