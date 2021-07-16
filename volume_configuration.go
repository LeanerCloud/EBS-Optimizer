package main

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type volumeConfig struct {
	VolumeType types.VolumeType
	IOPS       int32
	Throughput int32
	Region     string
	Size       int32
}

func (vc *volumeConfig) calculateMonthlyPrice() float64 {
	vi := ebsInfo[string(vc.VolumeType)]
	rp := vi.Pricing[vc.Region]

	debug.Printf("Calculating monthly cost for %v in %s \n", vc, vc.Region)

	// Start with Storage pricing

	monthlyPrice := rp.pricePerGB * float64(vc.Size)

	// Add provisioned IOPS pricing
	for _, iopsMonthlyPrice := range rp.piopsPrices {
		if vc.IOPS >= iopsMonthlyPrice.endRange {
			monthlyPrice += iopsMonthlyPrice.pricePerPIOPS * float64(iopsMonthlyPrice.endRange-iopsMonthlyPrice.beginRange)
		} else if vc.IOPS >= iopsMonthlyPrice.beginRange {
			monthlyPrice += iopsMonthlyPrice.pricePerPIOPS * float64(vc.IOPS-iopsMonthlyPrice.beginRange)
		}
	}

	// Add provisioned Throughput pricing
	for _, tputMonthlyPrice := range rp.tputPrices {
		if vc.Throughput >= tputMonthlyPrice.endRange {
			monthlyPrice += tputMonthlyPrice.tputPricePerMBps * float64(tputMonthlyPrice.endRange-tputMonthlyPrice.beginRange)
		} else if vc.Throughput > tputMonthlyPrice.beginRange {
			monthlyPrice += tputMonthlyPrice.tputPricePerMBps * float64(vc.Throughput-tputMonthlyPrice.beginRange)
		}
	}
	log.Printf("Cost for %#+v in %s: %f \n", vc, vc.Region, monthlyPrice)
	return monthlyPrice
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
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-north-1",
		"ap-east-1",
		"ap-south-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ap-northeast-2",
		"me-south-1",
	}
	for _, item := range io2SupporedRegions {
		if item == region {
			return true
		}
	}
	return false
}
