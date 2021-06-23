package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type region struct {
	name string

	conf *Config

	api ec2Conn

	ebsVolumes []*EBSVolume
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

func (r *region) processRegion() {

	log.Println("Creating connections to the required AWS services in", r.name)
	r.api.connect(r.name, r.conf.MainRegion)

	if !r.hasEBSPricingInfo() {
		r.getEBSPricingInfo()
	}

	r.scanEBSVolumes()
	r.processEBSVolumes()
}

func (r *region) scanEBSVolumes() error {
	resp, err := r.api.ec2.DescribeVolumes(&ec2.DescribeVolumesInput{})

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

//todo implement hasEBSPricingInfo
func (r *region) hasEBSPricingInfo() bool {
	return false
}

//todo implement getEBSPricingInfo
func (r *region) getEBSPricingInfo() error {
	return nil
}
