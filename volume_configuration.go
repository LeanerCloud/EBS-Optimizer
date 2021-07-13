package main

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

type volumeConfig struct {
	VolumeType   string
	IOPS         int64
	Throughput   int64
	monthlyPrice float64
}

func (vc *volumeConfig) calculatePrice() float64 {
	return 0
}

func (vc *volumeConfig) toString() string {
	res, err := json.Marshal(vc)
	if err != nil {
		log.Printf("error converting configuration %v to string: %v\n", vc, err.Error())
		return ""
	}
	debug.Printf("marshaled configuration: %v to %v\n", vc, res)
	return string(res)
}

func io2Supports(region string) bool {
	// this is fugly, I wish we had a better way of checking this.
	io2SupporedRegions := []string{
		endpoints.UsEast1RegionID,
		endpoints.UsEast2RegionID,
		endpoints.UsWest1RegionID,
		endpoints.UsWest2RegionID,
		endpoints.CaCentral1RegionID,
		endpoints.EuCentral1RegionID,
		endpoints.EuWest1RegionID,
		endpoints.EuWest2RegionID,
		endpoints.EuNorth1RegionID,
		endpoints.ApEast1RegionID,
		endpoints.ApSouth1RegionID,
		endpoints.ApSoutheast1RegionID,
		endpoints.ApSoutheast2RegionID,
		endpoints.ApNortheast1RegionID,
		endpoints.ApNortheast2RegionID,
		endpoints.MeSouth1RegionID,
	}
	for _, item := range io2SupporedRegions {
		if item == region {
			return true
		}
	}
	return false
}
