package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type region struct {
	name string

	conf *Config

	api ec2Conn

	ebsVolumes []*EBSVolume
	savings    float64
}

// var regionMap = map[string]string{
// 	"ap-east-1":      "Asia Pacific (Hong Kong)",
// 	"ap-northeast-1": "Asia Pacific (Tokyo)",
// 	"ap-northeast-2": "Asia Pacific (Seoul)",
// 	"ap-south-1":     "Asia Pacific (Mumbai)",
// 	"ap-southeast-1": "Asia Pacific (Singapore)",
// 	"ap-southeast-2": "Asia Pacific (Sydney)",
// 	"ca-central-1":   "Canada (Central)",
// 	"eu-central-1":   "EU (Frankfurt)",
// 	"eu-north-1":     "EU (Stockholm)",
// 	"eu-west-1":      "EU (Ireland)",
// 	"eu-west-2":      "EU (London)",
// 	"eu-west-3":      "EU (Paris)",
// 	"me-south-1":     "Middle East (Bahrain)",
// 	"sa-east-1":      "South America (Sao Paulo)",
// 	"us-east-1":      "US East (N. Virginia)",
// 	"us-east-2":      "US East (Ohio)",
// 	"us-west-1":      "US West (N. California)",
// 	"us-west-2":      "US West (Oregon)",
// }

var reverseRegionMap = map[string]string{
	"Asia Pacific (Hong Kong)":  "ap-east-1",
	"Asia Pacific (Tokyo)":      "ap-northeast-1",
	"Asia Pacific (Seoul)":      "ap-northeast-2",
	"Asia Pacific (Mumbai)":     "ap-south-1",
	"Asia Pacific (Singapore)":  "ap-southeast-1",
	"Asia Pacific (Sydney)":     "ap-southeast-2",
	"Canada (Central)":          "ca-central-1",
	"EU (Frankfurt)":            "eu-central-1",
	"EU (Stockholm)":            "eu-north-1",
	"EU (Ireland)":              "eu-west-1",
	"EU (London)":               "eu-west-2",
	"EU (Paris)":                "eu-west-3",
	"Middle East (Bahrain)":     "me-south-1",
	"South America (Sao Paulo)": "sa-east-1",
	"US East (N. Virginia)":     "us-east-1",
	"US East (Ohio)":            "us-east-2",
	"US West (N. California)":   "us-west-1",
	"US West (Oregon)":          "us-west-2",
}

func getRegion(regionDescription string) string {
	r := reverseRegionMap[regionDescription]
	if r == "" {
		return regionDescription
	}
	return r
}

func (r *region) enabled() bool {

	var enabledRegions []string

	if r.conf.Regions != "" {
		// Allow both space- and comma-separated values for the region list.
		csv := strings.Replace(r.conf.Regions, " ", ",", -1)
		enabledRegions = strings.Split(csv, ",")
	} else {
		return true
	}

	for _, region := range enabledRegions {

		// glob matching for region names
		if match, _ := filepath.Match(region, r.name); match {
			return true
		}
	}

	return false
}

func (r *region) scanEBSVolumes() error {
	resp, err := r.api.ec2.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})

	if err != nil {
		log.Println("Could not scan volumes", err.Error())
		return err
	}

	for _, v := range resp.Volumes {
		r.ebsVolumes = append(r.ebsVolumes, &EBSVolume{Volume: v, api: r.api, region: r.name})
	}
	return nil
}

func (r *region) processEBSVolumes() error {
	for _, v := range r.ebsVolumes {
		err := v.process()
		if err != nil {
			log.Println("Could not convert volume", v, err.Error())
			return err
		}
	}
	return nil
}

func (r *region) calculateHourlySavings() {
	var savings float64
	for _, v := range r.ebsVolumes {
		savings += v.calculateHourlySavings()
	}
	r.savings = savings
}
