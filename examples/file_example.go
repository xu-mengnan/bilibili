package main

import (
	"fmt"
	"log"

	"bilibili/pkg/file"
)

func main() {
	// 示例数据
	data := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "25", "New York"},
		{"Bob", "30", "San Francisco"},
		{"Charlie", "35", "Los Angeles"},
	}

	// 写入Excel文件
	fmt.Println("Writing Excel file...")
	err := file.WriteExcel(data, "example.xlsx")
	if err != nil {
		log.Fatal("Failed to write Excel file:", err)
	}
	fmt.Println("Excel file written successfully!")

	// 读取Excel文件
	fmt.Println("Reading Excel file...")
	excelData, err := file.ReadExcel("example.xlsx")
	if err != nil {
		log.Fatal("Failed to read Excel file:", err)
	}
	fmt.Println("Excel data:")
	for _, row := range excelData {
		fmt.Println(row)
	}

	// 写入CSV文件
	fmt.Println("Writing CSV file...")
	err = file.WriteCSV(data, "example.csv")
	if err != nil {
		log.Fatal("Failed to write CSV file:", err)
	}
	fmt.Println("CSV file written successfully!")

	// 读取CSV文件
	fmt.Println("Reading CSV file...")
	csvData, err := file.ReadCSV("example.csv")
	if err != nil {
		log.Fatal("Failed to read CSV file:", err)
	}
	fmt.Println("CSV data:")
	for _, row := range csvData {
		fmt.Println(row)
	}
}