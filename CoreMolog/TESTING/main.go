package main

import (
	"fmt"
	"strings"
)

func main() {
	strParam := ""
	strValue := ""
	strOperator := ""

	strInput := "status=success,function=hitter"
	arrInput := strings.Split(strInput, ",")

	for x:=0; x<len(arrInput); x++ {
		arrInputX := arrInput[x]

		if strings.Contains(arrInputX, "=") {
			arrData := strings.Split(arrInputX, "=")
			strParam = "\"" + arrData[0] + "\""
			strValue = "\"" + arrData[1] + "\""
			strOperator = "="
		} else if strings.Contains(arrInputX, "<>") {
			arrData := strings.Split(arrInputX, "<>")
			strParam = "\"" + arrData[0] + "\""
			strValue = "\"" + arrData[1] + "\""
			strOperator = "<>"
		}
		fmt.Println(strParam + strOperator + strValue)
	}




}
