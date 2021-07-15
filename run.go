package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Init initializes some data structures reusable across multiple event runs
func (e *EBSOptimizer) Init(cfg *Config) {
	e.config = cfg

	e.config.setupLogging()
	// use this only to list all the other regions

	debug.Println("Connecting to EC2 in the main region", e.config.MainRegion)
	e.mainEC2Conn = e.connectEC2(e.config.MainRegion)
	eo = e

	err := populateEBSPricing()

	if err != nil {
		log.Fatalf("failed to get EBS pricing information: %v", err)
	}
}

func (e *EBSOptimizer) connectEC2(region string) *ec2.EC2 {

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	return ec2.New(sess,
		aws.NewConfig().WithRegion(region))
}

func (e *EBSOptimizer) run(event *json.RawMessage) {
	allRegions, err := e.getRegions()

	if err != nil {
		log.Println(err.Error())
		return
	}

	e.processRegions(allRegions)

	// Print Final Recap
	log.Println("####### BEGIN FINAL RECAP #######")
	for r, a := range e.config.FinalRecap {
		for _, t := range a {
			log.Printf("%s %s\n", r, t)
		}
	}
}

// getRegions generates a list of AWS regions.
func (e *EBSOptimizer) getRegions() ([]string, error) {
	var output []string

	log.Println("Scanning for available AWS regions")

	resp, err := e.mainEC2Conn.DescribeRegions(&ec2.DescribeRegionsInput{})

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	debug.Println(resp)

	for _, r := range resp.Regions {

		if r != nil && r.RegionName != nil {
			debug.Println("Found region", *r.RegionName)
			output = append(output, *r.RegionName)
		}
	}
	return output, nil
}

// processAllRegions iterates all regions in parallel, and replaces instances
// for each of the ASGs tagged with tags as specified by slice represented by cfg.FilterByTags
// by default this is all asg with the tag 'spot-enabled=true'.
func (e *EBSOptimizer) processRegions(regions []string) {
	var wg sync.WaitGroup

	for _, r := range regions {

		wg.Add(1)

		r := region{name: r, conf: e.config}

		go func() {

			if r.enabled() {
				log.Printf("Enabled to run in %s, processing region.\n", r.name)
				r.processRegion()
			} else {
				debug.Println("Not enabled to run in", r.name)
				debug.Println("List of enabled regions:", r.conf.Regions)
			}

			wg.Done()
		}()
	}
	wg.Wait()
}
