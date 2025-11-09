package file

import (
	"github.com/xuri/excelize/v2"
)

// WriteExcel 写入Excel文件
func WriteExcel(data [][]string, filePath string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// 创建一个工作表
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return err
	}

	// 写入数据
	for i, row := range data {
		for j, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				return err
			}
			f.SetCellValue("Sheet1", cellName, cell)
		}
	}

	// 设置默认工作表
	f.SetActiveSheet(index)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return err
	}

	return nil
}

// ReadExcel 读取Excel文件
func ReadExcel(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// 获取工作表中的所有行
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	return rows, nil
}