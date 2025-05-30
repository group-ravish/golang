package reader

import (
	"encoding/csv"
	"errors"
	"os"
)

var osFileOpenFunc = os.Open

func isFile(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func ReadCSV(filePath string, lineNum int) (any, error) {

	isFile, err := isFile(filePath)
	if err != nil {
		return nil, err
	}
	if !isFile {
		return nil, errors.New("the provided path is a directory")
	}

	file, err := osFileOpenFunc(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errors.New("csv file is empty")
	}

	data := make(map[int][]string)
	for i, row := range records {
		data[i+1] = row
	}

	if lineNum == -1 {
		return data, nil
	}

	if row, exists := data[lineNum]; exists {
		return row, nil
	}

	return nil, errors.New("line number out of range")
}
