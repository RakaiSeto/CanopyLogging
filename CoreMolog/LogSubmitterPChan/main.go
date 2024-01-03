package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/ratelimit"
	"molog/modules"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type logData struct {
	Datetime  		string		`bson:"datetime"`
	LogLevel  		string		`bson:"loglevel"`
	TraceID 		string    	`bson:"traceid"`
	Application 	string    	`bson:"application"`
	Module 			string    	`bson:"module"`
	Function 		string    	`bson:"function"`
	Identity 		string    	`bson:"identity"`
	RemoteIP 		string    	`bson:"remoteip"`
	Message 		string    	`bson:"message"`
	ErrorMessage 	string    	`bson:"errormessage"`
	Status 			string    	`bson:"status"`
}

var dbM *mongo.Client
var cxM context.Context

var connRabbit *amqp.Connection
var channelLog *amqp.Channel
var queueLog = "TRCV_MONGO_LOG"
var tps = 10
var qosCount = 5


func doInsertToMongo(db *mongo.Client, ctx context.Context, dbLog string, incLogLevel string, incTraceID string, incApplication string,
	incModule string, incFunction string, incIdentity string, incRemoteIP string, incMessage string, isError bool, incErrorMessage error) {

	strDatetime := modules.DoFormatDateTime("YYYY-0M-0D HH:mm:ss.S", time.Now())
	colDatetime := "log_" + modules.DoFormatDateTime("YYYY0M0D", time.Now())
	strStatus := "SUCCESS"
	strErrorMessage := ""

	if isError {
		strStatus = "FAILED"
	}

	if incErrorMessage == nil {
		strErrorMessage = ""
	} else {
		strErrorMessage = fmt.Sprintf("%s", incErrorMessage)
	}

	_, err := db.Database(dbLog).Collection(colDatetime).InsertOne(ctx, logData{strDatetime, incLogLevel, incTraceID, incApplication,
		incModule, incFunction, incIdentity, incRemoteIP, incMessage, strErrorMessage, strStatus})
	if err != nil {
		modules.DoLog("ERROR", "LOG SERVER", "DOLOG", "Save",
			"Failed to save Logging. Error occurred.", true, err)
	}
}

func processTheMessage(queueMessage string) {

	mapQueueMessage := modules.ConvertJSONStringToMap("", queueMessage)

	traceId := modules.GetStringFromMapInterface(mapQueueMessage, "traceid")
	application := modules.GetStringFromMapInterface(mapQueueMessage, "application")
	module := modules.GetStringFromMapInterface(mapQueueMessage, "module")
	function := modules.GetStringFromMapInterface(mapQueueMessage, "function")
	identity := modules.GetStringFromMapInterface(mapQueueMessage, "identity")
	remoteip := modules.GetStringFromMapInterface(mapQueueMessage, "remoteip")
	message := modules.GetStringFromMapInterface(mapQueueMessage, "message")
	strError := modules.GetStringFromMapInterface(mapQueueMessage, "errormessage")

	isError := true
	incErrorStatus := "ERROR"

	if strings.Contains(strError, "nil") {
		strError = ""
	}
	incError := errors.New(strError)

	if len(strError) == 0 {
		isError = false
		incErrorStatus = "INFO"
	}

	incDbLog := strings.Replace(application, " ", "", -1)
	incDbLog = strings.Replace(incDbLog, ".", "", -1)
	incDbLog = strings.Replace(incDbLog, "-", "", -1)
	incDbLog = strings.ToLower(incDbLog)


	doInsertToMongo(dbM, cxM, incDbLog, incErrorStatus, traceId, application, module, function, identity, remoteip, message, isError, incError)
}

func processQueue() {

	theRateLimit := ratelimit.New(tps)

	queueIncomingTransmitter, errIncomingMitracomm := channelLog.QueueDeclare(
		queueLog,
		true,
		false,
		false,
		false,
		nil,
	)

	if errIncomingMitracomm != nil {
		modules.DoLog("INFO", "", "Transceiver9POINTS", "processQueue",
			"Failed to connect to queue "+queueLog+". Error occured.", true, errIncomingMitracomm)

		panic(errIncomingMitracomm)
	}

	errQ := channelLog.Qos(qosCount, 0, false)
	if errQ != nil {
		modules.DoLog("INFO", "", "Transceiver9POINTS", "processQueue",
			"Failed to make rabbitmq QOS to "+strconv.Itoa(qosCount), true, errQ)

		panic(errQ)
	}

	// consume
	messageTransmitter, err := channelLog.Consume(
		queueIncomingTransmitter.Name, // queue
		"",                            // consumer
		false,                         // auto-ack
		false,                         // exclusive
		false,                         // no-local
		false,                         // no-wait
		nil,                           // args
	)
	if err != nil {
		modules.DoLog("INFO", "", "Transceiver9POINTS", "processQueue",
			"Failed to consume queue "+queueLog+". Error occured.", true, err)
	} else {
		forever := make(chan bool)

		for d := range messageTransmitter {
			modules.DoLog("INFO", "", "Transceiver9POINTS", "processQueue",
				"Receiving message: "+string(d.Body), false, nil)

			// do the process with rateLimit transaction per second
			modules.DoLog("INFO", "", "Transceiver9POINTS", "processQueue",
				"Do processing the incoming message from queue "+queueLog, false, nil)
			theRateLimit.Take()
			theRateLimit.Take()
			theRateLimit.Take()
			theRateLimit.Take()
			theRateLimit.Take()

			queueMessage := string(d.Body)
			go processTheMessage(queueMessage)

			errx := d.Ack(false)

			if errx != nil {
				modules.DoLog("DEBUG", "", "Transceiver9POINTS", "readQueue",
					"Failed to acknowledge manually message: "+string(d.Body)+". STOP the transceiver.", false, nil)

				os.Exit(-1)
			}

			modules.DoLog("DEBUG", "", "Transceiver9POINTS", "readQueue",
				"Done Processing queue message "+string(d.Body)+". Sending ack to rabbitmq.", false, nil)
		}

		fmt.Println("[*] Waiting for messages. To exit press CTRL-C")
		<-forever
	}
}

func startReceiver() {

	var errI error
	channelLog, errI = connRabbit.Channel()
	if errI != nil {
		modules.DoLog("INFO", "", "Transceiver9POINTS", "doConnect",
			"Failed to create channel to rabbitmq to read incoming queue: "+queueLog, true, errI)

		panic(errI)
	}
	defer channelLog.Close()

	// Thread utk check status chIncoming and reconnect if failed. Do it in different treads run forever
	go func() {
		for {
			c := make(chan int)

			theCheckedQueue, errC := channelLog.QueueInspect(queueLog)

			if errC != nil {
				modules.DoLog("INFO", "", "Transceiver9POINTS", "doConnect",
					"Checking queue status "+queueLog+" is failed for error. DO RE-INITIATE INCOMING CHANNEL.", true, errC)

				channelLog, _ = connRabbit.Channel()
			} else {
				if theCheckedQueue.Consumers == 0 {
					modules.DoLog("INFO", "", "Transceiver9POINTS", "doConnect",
						"Consumer of the incoming queue: "+queueLog+" is 0. DO RE-INITIATE RABBITMQ CHANNEL.", false, nil)

					channelLog, _ = connRabbit.Channel()
					_ = channelLog.Qos(qosCount, 0, false)
				}
			}

			sleepDuration := 10 * time.Minute
			time.Sleep(sleepDuration)

			<-c
		}
	}()

	processQueue()
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
	opts1 := options.Client().SetMaxPoolSize(50)
	opts2 := options.Client().SetMaxConnecting(10)
	dbM, errM = mongo.Connect(cxM, opts, opts1, opts2)
	if errM != nil {
		panic(errM)
	}
	defer dbM.Disconnect(cxM)
	if errM1 = dbM.Ping(cxM, readpref.Primary()); errM1 != nil {
		panic(errM1)
	}


	// Initiate RabbitMQ
	var errRabbit error
	connRabbit, errRabbit = amqp.Dial("amqp://" + modules.MapConfig["rabbitUser"] + ":" + modules.MapConfig["rabbitPass"] + "@" + modules.MapConfig["rabbitHost"] + ":" + modules.MapConfig["rabbitPort"] + "/" + modules.MapConfig["rabbitVHost"])
	if errRabbit != nil {
		modules.DoLog("INFO", "", "SMPP20", "main",
			"Failed to connect to RabbitMQ server. Error", true, errRabbit)
	} else {
		modules.DoLog("INFO", "", "SMPP20", "main",
			"Success to connect to RabbitMQ server.", false, nil)
	}


	startReceiver()
}