package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"molog/modules"
	"net/http"
	"runtime"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var dbM *mongo.Client
var cxM context.Context


func GetIPAddress(incRemoteAddress string) string {
	strIPAddress := strings.Split(incRemoteAddress, ":")
	return strIPAddress[0]
}

func main() {
	// Load configuration file
	modules.InitiateGlobalVariables(false)
	runtime.GOMAXPROCS(4)

	// Mongo Log Database
	var errM error
	var errM1 error
	cxM = context.TODO()
	mongoInfo := fmt.Sprintf("mongodb://%s:%s",modules.MapConfig["mongoDBHost"],modules.MapConfig["mongoDBPort"])
	opts := options.Client().ApplyURI(mongoInfo)
	opts2 := options.Client().SetMaxPoolSize(50)
	dbM, errM = mongo.Connect(cxM, opts, opts2)
	if errM != nil {
		panic(errM)
	}
	defer dbM.Disconnect(cxM)
	if errM1 = dbM.Ping(cxM, readpref.Primary()); errM1 != nil {
		panic(errM1)
	}

	// APITransaction API
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-type", "application/json")

		var bodyBytes []byte
		//var tracecodeX string
		var responseContent string

		incURL := fmt.Sprintf("%s", r.URL)[1:]

		if r.Body != nil && r.Method == "POST" {
			bodyBytes, _ = io.ReadAll(r.Body)

			//fmt.Println(bodyBytes)
			var incomingBody string
			incomingBody = string(bodyBytes)
			incomingBody = strings.Replace(incomingBody, "\t", "", -1)
			incomingBody = strings.Replace(incomingBody, "\n", "", -1)
			incomingBody = strings.Replace(incomingBody, "\r", "", -1)

			// Write back the buffer to Body context, so it can be used by later process
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			strDateTime := modules.DoFormatDateTime("YYYY-0M-0D HH:mm:ss.S", time.Now())
			//remoteIPAddress := GetIPAddress(r.RemoteAddr)

			var incomingHeader = make(map[string]interface{})
			incomingHeader["Content-Type"] = r.Header.Get("Content-Type")

			mapIncoming := modules.ConvertJSONStringToMap("", incomingBody)

			incTracecode := modules.GetStringFromMapInterface(mapIncoming, "tracecode")
			incApp := modules.GetStringFromMapInterface(mapIncoming, "application")
			incModule := modules.GetStringFromMapInterface(mapIncoming, "module")
			incFunction := modules.GetStringFromMapInterface(mapIncoming, "function")
			incIdentity := modules.GetStringFromMapInterface(mapIncoming, "identity")
			incRemoteIP := modules.GetStringFromMapInterface(mapIncoming, "remoteip")
			incMessage := modules.GetStringFromMapInterface(mapIncoming, "message")
			strError := modules.GetStringFromMapInterface(mapIncoming, "error")

			// Route the request
			if incURL == "log" &&
				len(incTracecode) > 0 && len(incApp) > 0 && len(incModule) > 0 && len(incFunction) > 0  {

				isError := true
				incErrorStatus := "ERROR"
				incShowingStatus := "ERROR"

				if strings.Contains(strError, "nil") {
					strError = ""
				}
				incError := errors.New(strError)

				if len(strError) == 0 {
					isError = false
					incErrorStatus = "INFO"
					incShowingStatus = "SUCCESS"
				}

				if len(incIdentity) == 0 {
					incIdentity = "Unknown"
				}

				incDbLog := strings.Replace(incApp, " ", "", -1)
				incDbLog = strings.Replace(incDbLog, ".", "", -1)
				incDbLog = strings.Replace(incDbLog, "-", "", -1)
				incDbLog = strings.ToLower(incDbLog)

				modules.DoLogDB(dbM, cxM, incDbLog, incErrorStatus, incTracecode, incApp,
					incModule, incFunction, incIdentity, incRemoteIP, incMessage, isError, incError)

				message := modules.TrimStringLength(strings.ToUpper(strDateTime), 25) + "   " + modules.TrimStringLength(incTracecode, 35) + "   " + modules.TrimStringLength(incRemoteIP, 15) + "   " +  modules.TrimStringLength(strings.ToUpper(incShowingStatus), 10) + "   " + strings.ToUpper(incApp)
				fmt.Println(message)

			}
			//modules.SaveIncomingResponse(db, tracecodeX, responseHeader, responseContent)

			responseHeader := make(map[string]string)
			responseHeader["Content-Type"] = "application/json"

			mapResponse := make(map[string]interface{})
			mapResponse["status"] = "OK"

			responseContent = modules.ConvertMapInterfaceToJSON(mapResponse)
		}

		w.Write([]byte(responseContent))
	})

	thePort := "55555"
	log.Println("Starting HTTP Logging Server at port " + thePort)
	http.ListenAndServe(":" + thePort, nil)


}
