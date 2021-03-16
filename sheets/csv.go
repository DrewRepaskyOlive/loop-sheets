package sheets

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type SheetContent struct {
	Headers []string
	Content []map[string]string
}

func ReadCSV(filePath string) (*SheetContent, error) {
	headers, returnMap, err := csvFileToMap(filePath)
	if err != nil {
		return &SheetContent{}, err
	}
	return &SheetContent{
		Headers: headers,
		Content: returnMap,
	}, nil
}

// csvFileToMap reads csv file into slice of map
// slice is the line number
// map[string]string where key is column name
// ref: https://stackoverflow.com/a/57102302
func csvFileToMap(filePath string) (header []string, returnMap []map[string]string, err error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf(err.Error())
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	rawCSVData, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read CSV from file %q: %w", filePath, err)
	}

	header = []string{} // holds first row (header)
	for lineNum, record := range rawCSVData {
		// for first row, build the header slice
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				header = append(header, strings.TrimSpace(record[i]))
			}
		} else {
			// for each cell, map[string]string k=header v=value
			line := map[string]string{}
			for i := 0; i < len(record); i++ {
				line[header[i]] = record[i]
			}
			returnMap = append(returnMap, line)
		}
	}

	return
}
