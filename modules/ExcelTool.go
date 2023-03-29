package modules

import (
	"fmt"
)

func GetRowColByName(rows [][]string, rowFirstColName string, colHeaderName string) string {
	fmt.Println("rowFirstColName: " + rowFirstColName + ", colHeaderName: " + colHeaderName)
	DoLog("DEBUG", "", "ExcelTool", "GetRowColByName",
		"Getting value of rowFirstColName: "+rowFirstColName+", colHeaderName: "+colHeaderName, false, nil)

	selectedRow := -1
	selectedCol := -1
	selectedVal := ""

	noOfRow := len(rows)
	for x := 0; x < noOfRow; x++ {
		fmt.Printf("%d. Jumlah column di row %d adalah: %d\n", x, x, len(rows[x]))

		noOfColumn := len(rows[x])
		for y := 0; y < noOfColumn; y++ {
			fmt.Printf("%d - %d. value: %s\n", x, y, rows[x][y])

			// Checkin row
			if y == 0 && rows[x][y] == rowFirstColName {
				selectedRow = x
			}

			// Checkin col
			if x == 0 && rows[x][y] == colHeaderName {
				selectedCol = y
			}
		}
	}

	if selectedRow > -1 && selectedCol > -1 {
		selectedVal = rows[selectedRow][selectedCol]
		DoLog("DEBUG", "", "ExcelTool", "GetRowColByName",
			"Getting value of rowFirstColName: "+rowFirstColName+", colHeaderName: "+colHeaderName+" -> value: "+selectedVal, false, nil)
	} else {
		selectedVal = ""
		DoLog("DEBUG", "", "ExcelTool", "GetRowColByName",
			"Failed to get content value for rowFirstColName: "+rowFirstColName+", colHeaderName: "+colHeaderName+". Data NOT FOUND.", false, nil)
	}

	return selectedVal
}
