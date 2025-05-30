package main

import (
	"flag"
	"fmt"

	"main.go/reader"
)

func main() {

	lineNum := flag.Int("lineNum", -1, "To read one or all lines from a csv file")

	flag.Parse()

	fmt.Println("Line to read:", *lineNum)

	val, err := reader.ReadCSV("sample/customers-100.csv", *lineNum)
	if err != nil {
		fmt.Printf("reader error > %s", err)
	}
	switch v := val.(type) {
	case [][]string:
		for i, row := range v {
			fmt.Printf("Content from line number %d is %v", i, row)
		}
	case []string:
		fmt.Println("Content in line is:", v)
	default:
		fmt.Println("Unkown Type", v)
	}
}
