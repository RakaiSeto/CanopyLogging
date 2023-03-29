package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"molog/modules"
	"net/http"
)

func HitLogPost(incTrace string, incApp string, incMod string, incFunc string, incMessage string, incError error) {

	url := modules.MapConfig["logEndpoint"]
	strError := fmt.Sprintf("%v", incError)

	mapRequest := make(map[string]interface{})
	mapRequest["tracecode"] = incTrace
	mapRequest["application"] = incApp
	mapRequest["module"] = incMod
	mapRequest["function"] = incFunc
	mapRequest["message"] = incMessage
	mapRequest["error"] = strError

	jsonRequest := modules.ConvertMapInterfaceToJSON(mapRequest)
	var bodyByte = []byte(jsonRequest)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(bodyByte))

	if err != nil {
		log.Fatal(err)
	}


	if len(strError) == 0 || strError == "nil" {
		modules.DoLog("INFO", incTrace, incMod, incFunc, incMessage, false, nil)
	} else {
		modules.DoLog("ERROR", incTrace, incMod, incFunc, incMessage, true, incError)
	}
}


func main() {
	modules.InitiateGlobalVariables(false)

	HitLogPost("ASSS", "monalisa", "RPC Admin", "Do Login", "Normal request", errors.New("error"))
}
