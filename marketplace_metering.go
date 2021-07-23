package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
)

const EBSOptimizerMarketplaceProductID = "a01frq6kyyl44oswtf79gz7rs"

func meterMarketplaceUsage(savings float64) error {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("Could not create Marketplace service connections")
	}

	svc := marketplacemetering.NewFromConfig(cfg)

	res, err := svc.MeterUsage(context.TODO(), &marketplacemetering.MeterUsageInput{
		ProductCode:    aws.String(EBSOptimizerMarketplaceProductID),
		Timestamp:      aws.Time(time.Now()),
		UsageDimension: aws.String("SavingsCut"),
		UsageQuantity:  aws.Int32(int32(savings * 1000 * 0.05)),
	})

	if err != nil {
		fmt.Printf("Error submitting Marketplace metering data: %v, received response: %v\n", err.Error(), res)
		return err
	}
	return nil
}
