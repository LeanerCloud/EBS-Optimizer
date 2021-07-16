package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (e *EBSOptimizer) connectEC2(region string) *ec2.Client {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		panic(err)
	}

	return ec2.NewFromConfig(cfg)
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

	debug.Println("Scanning for available AWS regions")

	resp, err := e.mainEC2Conn.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	debug.Println(resp)

	for _, r := range resp.Regions {

		if r.RegionName != nil {
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
	var mutex sync.Mutex
	var savings float64

	for _, reg := range regions {

		wg.Add(1)

		r := region{name: reg, conf: e.config}

		go func() {

			debug.Println("Creating connections to the required AWS services in", r.name)
			r.api.connect(r.name, r.conf.MainRegion)
			r.scanEBSVolumes()

			if r.enabled() {
				log.Printf("Enabled to run in %s, processing region.\n", r.name)
				r.processEBSVolumes()
			} else {
				debug.Println("Not enabled to run in", r.name)
				debug.Println("List of enabled regions:", r.conf.Regions)
			}

			r.calculateHourlySavings()
			if r.savings > 0 {
				log.Printf("Calculated savings in %s: $%f(monthly), %f(hourly) ", r.name, r.savings*730, r.savings)
			}

			mutex.Lock()
			savings += r.savings
			mutex.Unlock()

			wg.Done()
		}()
	}
	wg.Wait()

	log.Printf("Total savings: %f(monthly), %f(hourly)", savings*730, savings)
}
