package main

import (
	"io/ioutil"
	"log"
	"os"
)

func (cfg *Config) setupLogging() {

	cfg.LogFile = os.Stdout
	cfg.LogFlag = log.Ldate | log.Ltime | log.Lshortfile

	log.SetOutput(cfg.LogFile)
	log.SetFlags(cfg.LogFlag)

	if os.Getenv("EBS_OPTIMIZER_DEBUG") == "true" {
		debug = log.New(cfg.LogFile, "", cfg.LogFlag)
	} else {
		debug = log.New(ioutil.Discard, "", 0)
	}

}
