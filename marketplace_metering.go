package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const EBSOptimizerMarketplaceProductID = "a01frq6kyyl44oswtf79gz7rs"

//SSMParameterName stores the name of the SSM parameter that stores the success status of the latest metering call
const SSMParameterName = "ebs-optimizer-metering"

func meterMarketplaceUsage(savings float64) error {

	// Metering is supposed to be done from Fargate, but we check it here and return an error in case it failed before
	if runningFromLambda() {
		log.Println("Running from Lambda")
		if failedFromFargate() {
			log.Println("Metering failed previously, exiting...")
			return errors.New("metering previously failed")
		}
		log.Println("Metering succeeded previously from Fargate, moving on...")
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Printf("Could not create Marketplace service connections")
		return err
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
	markAsSuccessfulFromFargate()
	return nil
}

func putSSMParameter(status string) {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Printf("Could not create SSM service connections")
		return
	}

	svc := ssm.NewFromConfig(cfg)

	_, err = svc.PutParameter(context.TODO(), &ssm.PutParameterInput{
		Name:      aws.String(SSMParameterName),
		Overwrite: true,
		Type:      types.ParameterTypeString,
		Value:     aws.String(status),
	})
	if err != nil {
		log.Printf("Error persisting marketplace metering status(%s) to SSM: %s", status, err.Error())
	}
}

func markAsSuccessfulFromFargate() {
	putSSMParameter("success")
}

func markAsFailingFromFargate() {
	putSSMParameter("failure")
}

func failedFromFargate() bool {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Printf("Could not create SSM service connections")
		return true
	}

	svc := ssm.NewFromConfig(cfg)
	res, err := svc.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name: aws.String(SSMParameterName),
	})

	if err != nil {
		log.Printf("Error reading marketplace metering status from SSM")

		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			log.Printf("Parameter not found: %v", err.Error())
			return false
		}
		log.Printf("Encountered error: %v", err.Error())
		return true
	}

	status := *res.Parameter.Value
	log.Printf("Retrieved marketplace metering status from SSM: %s", status)
	return status == "failure"
}
