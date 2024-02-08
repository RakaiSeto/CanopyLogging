package main

import (
	"bytes"
	"canopyLogging/modules"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var connRabbit *amqp.Connection
var queueLog = "TRCV_MONGO_LOG"
var channelLog *amqp.Channel


func doSubmitToQueue(chIncoming *amqp.Channel, datetime string, loglevel string, traceid string, application string, module string,
	function string, identity string, remoteip string, message string, errormessage string, status string) {

	var mapLog = make(map[string]string)
	mapLog["datetime"] = datetime
	mapLog["loglevel"] = loglevel
	mapLog["traceid"] = traceid
	mapLog["application"] = application
	mapLog["module"] = module
	mapLog["function"] = function
	mapLog["identity"] = identity
	mapLog["remoteip"] = remoteip
	mapLog["message"] = message
	mapLog["errormessage"] = errormessage
	mapLog["status"] = status

	jsonMessage := modules.ConvertMapStringToJSON(mapLog)

	errP := chIncoming.Publish(
		"",            // exchange
		queueLog, // routing key
		false,         // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(jsonMessage),
		})

	if errP != nil {
		// Reconnect the channel and re-publish
		var errT error
		chIncoming, errT = connRabbit.Channel()
		if errT != nil {
			modules.DoLog("DEBUG", traceid, "Receiver", "doSubmitToQueue",
				"Failed to re-initiate channel rabbitMQ. Error occured.", true, errT)
		}

		_ = chIncoming.Publish(
			"",            // exchange
			queueLog, // routing key
			false,         // mandatory
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(jsonMessage),
			})
	} else {
		modules.DoLog("DEBUG", traceid, "canopyLogging", "Submit to Queue",
			"Successfully to publish to SMS INCOMING Queue " + queueLog, false, nil)
	}
	chIncoming.Close()
}


func GetIPAddress(incRemoteAddress string) string {
	strIPAddress := strings.Split(incRemoteAddress, ":")
	return strIPAddress[0]
}

func main() {
	// Load configuration file
	modules.InitiateGlobalVariables()
	runtime.GOMAXPROCS(4)

	// Initiate RabbitMQ
	var errRabbit error
	connRabbit, errRabbit = amqp.Dial("amqp://" + modules.MapConfig["rabbitUser"] + ":" + modules.MapConfig["rabbitPass"] + "@" + modules.MapConfig["rabbitHost"] + ":" + modules.MapConfig["rabbitPort"] + "/" + modules.MapConfig["rabbitVHost"])
	if errRabbit != nil {
		modules.DoLog("INFO", "LOG SERVER", "DOLOG", "main",
			"Failed to connect to RabbitMQ server. Error", true, errRabbit)
	} else {
		modules.DoLog("INFO", "LOG SERVER", "DOLOG", "main",
			"Success to connect to RabbitMQ server.", false, nil)
	}
	defer connRabbit.Close()

	

	// APITransaction API
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-type", "application/json")

		var bodyBytes []byte
		var responseContent string

		incURL := fmt.Sprintf("%s", r.URL)[1:]

		if r.Body != nil && r.Method == "POST" {
			bodyBytes, _ = io.ReadAll(r.Body)

			var incomingBody string
			incomingBody = string(bodyBytes)
			incomingBody = strings.Replace(incomingBody, "\t", "", -1)
			incomingBody = strings.Replace(incomingBody, "\n", "", -1)
			incomingBody = strings.Replace(incomingBody, "\r", "", -1)

			// Write back the buffer to Body context, so it can be used by later process
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			//strDateTime := modules.DoFormatDateTime("YYYY-0M-0D HH:mm:ss.S", time.Now())
			//remoteIPAddress := GetIPAddress(r.RemoteAddr)

			var incomingHeader = make(map[string]interface{})
			incomingHeader["Content-Type"] = r.Header.Get("Content-Type")
			r.Body.Close()

			mapIncoming := modules.ConvertJSONStringToMap("", incomingBody)

			incTracecode := modules.GetStringFromMapInterface(mapIncoming, "tracecode")
			incApp := modules.GetStringFromMapInterface(mapIncoming, "application")
			incModule := modules.GetStringFromMapInterface(mapIncoming, "module")
			incFunction := modules.GetStringFromMapInterface(mapIncoming, "function")
			incIdentity := modules.GetStringFromMapInterface(mapIncoming, "identity")
			incRemoteIP := modules.GetStringFromMapInterface(mapIncoming, "remoteip")
			incMessage := modules.GetStringFromMapInterface(mapIncoming, "message")
			incStatusCode := modules.GetStringFromMapInterface(mapIncoming, "code")
			strError := modules.GetStringFromMapInterface(mapIncoming, "error")

			// Route the request
			if incURL == "log" &&
				len(incTracecode) > 0 && len(incApp) > 0 && len(incModule) > 0 && len(incFunction) > 0  {

				isError := true
				incLogLevel := "ERROR"
				strStatus := ""
				//incShowingStatus := "ERROR"
				if len(incStatusCode) > 0 {
					strStatus = incStatusCode + " - SUCCESS"
				} else {
					strStatus = "SUCCESS"
				}
				strErrorMessage := ""

				if strings.Contains(strError, "nil") {
					strError = ""
				}
				incError := errors.New(strError)

				if len(strError) == 0 {
					isError = false
					incLogLevel = "INFO"
					//incShowingStatus = "SUCCESS"
				}

				if len(incIdentity) == 0 {
					incIdentity = "Unknown"
				}

				if incError == nil {
					strErrorMessage = ""
				} else {
					strErrorMessage = fmt.Sprintf("%s", incError)
				}

				if isError {
					if len(incStatusCode) > 0 {
						strStatus = incStatusCode + " - FAILED"
					} else {
						strStatus = "FAILED"
					}
				}

				incDbLog := strings.Replace(incApp, " ", "", -1)
				incDbLog = strings.Replace(incDbLog, ".", "", -1)
				incDbLog = strings.Replace(incDbLog, "-", "", -1)
				incDbLog = strings.ToLower(incDbLog)

				strDatetime := modules.DoFormatDateTime("YYYY-0M-0D HH:mm:ss.S", time.Now())

				doSubmitToQueue(channelLog, strDatetime, incLogLevel, incTracecode, incApp, incModule,
					incFunction, incIdentity, incRemoteIP, incMessage, strErrorMessage, strStatus)

			}

			responseHeader := make(map[string]string)
			responseHeader["Content-Type"] = "application/json"

			mapResponse := make(map[string]interface{})
			mapResponse["status"] = "OK"

			responseContent = modules.ConvertMapInterfaceToJSON(mapResponse)
		}

		w.Write([]byte(responseContent))

	})

	thePort := "55555"
	srv := &http.Server{
		Addr: ":" + thePort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Starting HTTP Logging Server at port " + thePort)
	srv.ListenAndServe()


}