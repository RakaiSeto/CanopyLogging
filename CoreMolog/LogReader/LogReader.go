package main

import (
	"bytes"
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"molog/modules"
	"net/http"
	"runtime"
	"strings"
)

var dbM *mongo.Client
var cxM context.Context

type LogData struct {
	Datetime		string
	LogLevel		string
	TraceID			string
	Application		string
	Modules			string
	Function		string
	Identity		string
	RemoteIP		string
	Message			string
	ErrorMessage	string
	Status			string
}

func GetIPAddress(incRemoteAddress string) string {
	strIPAddress := strings.Split(incRemoteAddress, ":")
	return strIPAddress[0]
}

func getAllData(incDatabase string, incCollection string) (Output []map[string]interface{}) {

	filter := bson.D{{}}
	opts := options.Find().SetSort(bson.D{{"_id", -1}})

	cursor, err := dbM.Database(incDatabase).Collection(incCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	var logData []LogData
	if err = cursor.All(context.TODO(), &logData); err != nil {
		log.Fatal(err)
	}

	results := []map[string]interface{}{}
	for _, value := range logData {
		result := make(map[string]interface{})
		result["datetime"] = value.Datetime
		result["traceid"] = value.TraceID
		result["loglevel"] = value.LogLevel
		result["application"] = value.Application
		result["module"] = value.Modules
		result["function"] = value.Function
		result["identity"] = value.Identity
		result["remoteip"] = value.RemoteIP
		result["message"] = value.Message
		result["error"] = value.ErrorMessage
		result["status"] = value.Status

		results = append(results, result)
	}

	return results
}

func findByTraceID(incDatabase string, incCollection string, incTraceID string) (Output []map[string]interface{}) {

	filter := bson.D{{"traceid", incTraceID}}
	opts := options.Find().SetSort(bson.D{{"_id", -1}})

	cursor, err := dbM.Database(incDatabase).Collection(incCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	var logData []LogData
	if err = cursor.All(context.TODO(), &logData); err != nil {
		log.Fatal(err)
	}

	results := []map[string]interface{}{}
	for _, value := range logData {
		result := make(map[string]interface{})
		result["datetime"] = value.Datetime
		result["traceid"] = value.TraceID
		result["loglevel"] = value.LogLevel
		result["application"] = value.Application
		result["module"] = value.Modules
		result["function"] = value.Function
		result["identity"] = value.Identity
		result["remoteip"] = value.RemoteIP
		result["message"] = value.Message
		result["error"] = value.ErrorMessage
		result["status"] = value.Status

		results = append(results, result)
	}

	return results
}

func findByRemoteIP(incDatabase string, incCollection string, incRemoteIP string) (Output []map[string]interface{}) {

	filter := bson.D{{"remoteip", incRemoteIP}}
	opts := options.Find().SetSort(bson.D{{"_id", -1}})

	cursor, err := dbM.Database(incDatabase).Collection(incCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	var logData []LogData
	if err = cursor.All(context.TODO(), &logData); err != nil {
		log.Fatal(err)
	}

	results := []map[string]interface{}{}
	for _, value := range logData {
		result := make(map[string]interface{})
		result["datetime"] = value.Datetime
		result["traceid"] = value.TraceID
		result["loglevel"] = value.LogLevel
		result["application"] = value.Application
		result["module"] = value.Modules
		result["function"] = value.Function
		result["identity"] = value.Identity
		result["remoteip"] = value.RemoteIP
		result["message"] = value.Message
		result["error"] = value.ErrorMessage
		result["status"] = value.Status

		results = append(results, result)
	}

	return results
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
	dbM, errM = mongo.Connect(cxM, opts)
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

		//var mapOutput := make(map[string]interface{})
		var bodyBytes []byte
		//var tracecodeX string
		var responseContent string

		incURL := fmt.Sprintf("%s", r.URL)[1:]

		if r.Body != nil && r.Method == "POST" {
			bodyBytes, _ = ioutil.ReadAll(r.Body)

			//fmt.Println(bodyBytes)
			var incomingBody string
			incomingBody = string(bodyBytes)
			incomingBody = strings.Replace(incomingBody, "\t", "", -1)
			incomingBody = strings.Replace(incomingBody, "\n", "", -1)
			incomingBody = strings.Replace(incomingBody, "\r", "", -1)

			// Write back the buffer to Body context, so it can be used by later process
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			//strDateTime := modules.DoFormatDateTime("YYYY-0M-0D HH:mm:ss.S", time.Now())
			//remoteIPAddress := GetIPAddress(r.RemoteAddr)

			mapOutput := []map[string]interface{}{}
			var incomingHeader = make(map[string]interface{})
			incomingHeader["Content-Type"] = r.Header.Get("Content-Type")

			mapIncoming := modules.ConvertJSONStringToMap("", incomingBody)

			incDatabase := modules.GetStringFromMapInterface(mapIncoming, "database")
			incDate := modules.GetStringFromMapInterface(mapIncoming, "date")
			incFilter := modules.GetStringFromMapInterface(mapIncoming, "filter")
			incFilterData := modules.GetStringFromMapInterface(mapIncoming, "data")

			// Route the request
			if incURL == "getlog" && len(incDatabase) > 0 && len(incDate) > 0  {

				incCollection := "log_" + incDate

				if strings.ToUpper(incFilter) == "TRACEID" && len(incFilterData) > 0 {
					mapOutput = findByTraceID(incDatabase, incCollection, incFilterData)
				} else if strings.ToUpper(incFilter) == "REMOTEIP" && len(incFilterData) > 0 {
					mapOutput = findByRemoteIP(incDatabase, incCollection, incFilterData)
				} else {
					mapOutput = getAllData(incDatabase, incCollection)
				}
			}
			//modules.SaveIncomingResponse(db, tracecodeX, responseHeader, responseContent)

			responseHeader := make(map[string]string)
			responseHeader["Content-Type"] = "application/json"

			mapResponse := make(map[string]interface{})
			mapResponse["results"] = mapOutput
			mapResponse["status"] = "OK"

			responseContent = modules.ConvertMapInterfaceToJSON(mapResponse)
		}

		w.Write([]byte(responseContent))
	})

	thePort := "55556"
	log.Println("Starting HTTP Logging Get Data Server at port " + thePort)
	http.ListenAndServe(":" + thePort, nil)
}
