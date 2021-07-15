package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go/aws"
)

type volumeInfo struct {
	name              string
	minSizeGB         int64
	maxSizeTB         int64
	minDurability     float64
	maxIOPS           int64
	maxIOPSPerGB      int64
	maxThroughput     int64
	Pricing           volumePricing
	throughputPerMBs  float64
	throughputFree    int64
	minIOPS           int64
	baselineIOPSPerGB int64
	iopsBurst         int64
	bootable          bool
}

type piopsPrice struct {
	beginRange, endRange int64
	pricePerPIOPS        float64
}

type tputPrice struct {
	beginRange, endRange int64
	tputPricePerMBps     float64
}

type regionalPricing struct {
	piopsPrices []piopsPrice
	tputPrices  []tputPrice
	pricePerGB  float64
}

type volumePricing map[string]regionalPricing

var ebsInfo = map[string]volumeInfo{
	"gp2": {
		name:              "gp2",
		minSizeGB:         1,
		maxSizeTB:         16,
		minDurability:     99.8,
		maxIOPS:           16000,
		maxThroughput:     250,
		Pricing:           make(volumePricing),
		minIOPS:           100,
		baselineIOPSPerGB: 3,
		iopsBurst:         3000,
		bootable:          true,
	},
	"gp3": {
		name:             "gp3",
		minSizeGB:        1,
		maxSizeTB:        16,
		minDurability:    99.8,
		maxIOPS:          16000,
		maxThroughput:    1000,
		Pricing:          make(volumePricing),
		throughputPerMBs: 0.04,
		throughputFree:   125,
		bootable:         true,
	},
	"io1": {
		name:          "io1",
		minSizeGB:     4,
		maxSizeTB:     16,
		minDurability: 99.8,
		maxIOPS:       64000,
		maxIOPSPerGB:  50,
		maxThroughput: 1000,
		Pricing:       make(volumePricing),
		bootable:      true,
	},
	"io2": {
		name:          "io2",
		minSizeGB:     4,
		maxSizeTB:     16,
		minDurability: 99.999,
		maxIOPS:       64000,
		maxIOPSPerGB:  500,
		maxThroughput: 1000,
		Pricing:       make(volumePricing),
		bootable:      true,
	},

	"st1": {
		name:              "st1",
		minSizeGB:         125,
		maxSizeTB:         16,
		minDurability:     99.8,
		maxIOPS:           500,
		maxThroughput:     250,
		Pricing:           make(volumePricing),
		minIOPS:           100,
		baselineIOPSPerGB: 3,
		iopsBurst:         3000,
	},
	"sc1": {
		name:          "sc1",
		minSizeGB:     125,
		maxSizeTB:     16,
		minDurability: 99.8,
		maxIOPS:       250,
		maxThroughput: 250,
		Pricing:       make(volumePricing),
		minIOPS:       0,
	},
	"standard": {
		name:          "standard",
		minSizeGB:     1,
		maxSizeTB:     1,
		minDurability: 99.8,
		maxIOPS:       200,
		maxThroughput: 90,
		Pricing:       make(volumePricing),
		minIOPS:       0,
	},
}

func populateEBSPricing() error {

	err := populateStoragePricing()
	if err != nil {
		log.Fatalf("failed to get storage pricing information: %v", err)
	}

	err = populateGP3PIOPSPricing()
	if err != nil {
		log.Fatalf("failed to get GP3 PIOPS pricing information: %v", err)
	}

	err = populateGP3PThroughputPricing()

	if err != nil {
		log.Fatalf("failed to get GP3 Throughput pricing information: %v", err)
	}

	err = populateIO1IOPSPricing()

	if err != nil {
		log.Fatalf("failed to get IO1 PIOPS pricing information: %v", err)
	}

	err = populateIO2IOPSPricing()

	if err != nil {
		log.Fatalf("failed to get IO2 PIOPS pricing information: %v", err)
	}

	return err
}

func populateStoragePricing() error {
	var f = []types.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Field: aws.String("productFamily"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("Storage"),
		},
	}
	pd, err := getPricingData(f)
	if err != nil {
		return err
	}

	for _, v := range pd {
		volumeType := v.Product.Attributes.VolumeAPIName
		region := getRegion(v.Product.Attributes.Location)
		pricePerGBStr := v.Terms.OnDemand.SKU.PriceDimensions.Dimension.PricePerUnit.USD
		pricePerGB, err := strconv.ParseFloat(pricePerGBStr, 64)

		if err != nil {
			fmt.Printf("failed to convert %s to float", pricePerGBStr)
		}
		fmt.Printf("%s: %s costs %f \n %#v\n\n", volumeType, region, pricePerGB, v)
		ebsInfo[volumeType].Pricing[region] = regionalPricing{
			pricePerGB: pricePerGB,
		}
	}

	return nil
}

func populateGP3PIOPSPricing() error {
	var f = []types.Filter{
		{
			Field: aws.String("volumeApiName"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("gp3"),
		},
		{
			Field: aws.String("productFamily"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("System Operation"),
		},
	}
	pd, err := getPricingData(f)
	if err != nil {
		return err
	}

	for _, v := range pd {
		volumeType := v.Product.Attributes.VolumeAPIName
		region := getRegion(v.Product.Attributes.Location)
		priceStr := v.Terms.OnDemand.SKU.PriceDimensions.Dimension.PricePerUnit.USD
		price, err := strconv.ParseFloat(priceStr, 64)

		if err != nil {
			fmt.Printf("failed to convert %s to float", priceStr)
		}
		fmt.Printf("%s: %s PIOPS costs %f \n %#v\n\n", volumeType, region, price, v)

		pl := ebsInfo[volumeType].Pricing[region]

		pl.piopsPrices = append(
			pl.piopsPrices, piopsPrice{
				beginRange:    3000,
				endRange:      16000,
				pricePerPIOPS: price,
			})
		ebsInfo[volumeType].Pricing[region] = pl

	}
	return nil
}

func populateGP3PThroughputPricing() error {
	var f = []types.Filter{
		{
			Field: aws.String("volumeApiName"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("gp3"),
		},
		{
			Field: aws.String("productFamily"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("Provisioned Throughput"),
		},
	}
	pd, err := getPricingData(f)
	if err != nil {
		return err
	}

	for _, v := range pd {
		volumeType := v.Product.Attributes.VolumeAPIName
		region := getRegion(v.Product.Attributes.Location)
		priceStr := v.Terms.OnDemand.SKU.PriceDimensions.Dimension.PricePerUnit.USD
		price, err := strconv.ParseFloat(priceStr, 64)
		price = price / 1024 // the API returns the value in GBps

		if err != nil {
			fmt.Printf("failed to convert %s to float", priceStr)
		}
		fmt.Printf("%s: %s Throughput costs %f \n %#v\n\n", volumeType, region, price, v)

		pl := ebsInfo[volumeType].Pricing[region]

		pl.tputPrices = append(
			pl.tputPrices, tputPrice{
				beginRange:       125,
				endRange:         1000,
				tputPricePerMBps: price,
			})
		ebsInfo[volumeType].Pricing[region] = pl

	}
	return nil
}

func populateIO1IOPSPricing() error {
	var f = []types.Filter{
		{
			Field: aws.String("volumeApiName"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("io1"),
		},
		{
			Field: aws.String("productFamily"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("System Operation"),
		},
	}
	pd, err := getPricingData(f)
	if err != nil {
		return err
	}

	for _, v := range pd {
		volumeType := v.Product.Attributes.VolumeAPIName
		region := getRegion(v.Product.Attributes.Location)
		priceStr := v.Terms.OnDemand.SKU.PriceDimensions.Dimension.PricePerUnit.USD
		price, err := strconv.ParseFloat(priceStr, 64)

		if err != nil {
			fmt.Printf("failed to convert %s to float", priceStr)
		}
		fmt.Printf("%s: %s PIOPS costs %f \n %#v\n\n", volumeType, region, price, v)

		pl := ebsInfo[volumeType].Pricing[region]

		pl.piopsPrices = append(
			pl.piopsPrices, piopsPrice{
				beginRange:    0,
				endRange:      64000,
				pricePerPIOPS: price,
			})
		ebsInfo[volumeType].Pricing[region] = pl

	}
	return nil
}

func populateIO2IOPSPricing() error {
	var f = []types.Filter{
		{
			Field: aws.String("volumeApiName"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("io2"),
		},
		{
			Field: aws.String("productFamily"),
			Type:  types.FilterType("TERM_MATCH"),
			Value: aws.String("System Operation"),
		},
	}
	pd, err := getPricingData(f)
	if err != nil {
		return err
	}

	for _, v := range pd {
		var beginRange, endRange int64
		volumeType := v.Product.Attributes.VolumeAPIName
		region := getRegion(v.Product.Attributes.Location)
		priceStr := v.Terms.OnDemand.SKU.PriceDimensions.Dimension.PricePerUnit.USD
		price, err := strconv.ParseFloat(priceStr, 64)
		group := v.Product.Attributes.Group

		if err != nil {
			fmt.Printf("failed to convert %s to float", priceStr)
		}
		fmt.Printf("%s: %s PIOPS costs %f \n %#v\n\n", volumeType, region, price, v)

		pl := ebsInfo[volumeType].Pricing[region]

		if group == "EBS IOPS Tier 3" {
			beginRange = 64001
			endRange = 256000
		} else if group == "EBS IOPS Tier 2" {
			beginRange = 32001
			endRange = 64000
		} else if group == "EBS IOPS" {
			endRange = 32000
		}

		pl.piopsPrices = append(
			pl.piopsPrices, piopsPrice{
				beginRange:    beginRange,
				endRange:      endRange,
				pricePerPIOPS: price,
			})
		ebsInfo[volumeType].Pricing[region] = pl

	}
	return nil
}
