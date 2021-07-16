package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var conf Config

// Version represents the build version being used
var Version = "missing version string"

// ExpirationDate represents the date at which the version will expire
var ExpirationDate = "01-Jan-2100"

var eo *EBSOptimizer

var debug *log.Logger

// EBSOptimizer provides the global configuration
type EBSOptimizer struct {
	config      *Config
	mainEC2Conn *ec2.Client
}

func main() {

	eventFile := conf.EventFile

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
	} else if eventFile != "" {
		parseEvent, err := ioutil.ReadFile(eventFile)
		if err != nil {
			log.Fatal(err)
		}
		Handler(context.TODO(), parseEvent)
	} else {
		eventHandler(nil)
	}
}

func eventHandler(event *json.RawMessage) {

	log.Println("Starting ebs-optimizer, build ", Version)

	if isExpired(ExpirationDate) {
		log.Println("EBS-Optimizer expired, please install a newer version.")
		return
	}

	log.Printf("Configuration flags: %#v", conf)

	eo.run(event)
	log.Println("Execution completed, nothing left to do")
}

// this is the equivalent of a main for when running from Lambda, but on Lambda
// the runFromCronEvent() is executed within the handler function every time we have an event
func init() {
	conf = Config{Version: Version}
	conf.setupLogging()
	log.Println("Determining configuration")
	conf.ParseCommandlineFlags()

	eo = &EBSOptimizer{}

	eo.Init(&conf)
}

// Handler implements the AWS Lambda handler interface
func Handler(ctx context.Context, rawEvent json.RawMessage) {
	eventHandler(&rawEvent)
}
