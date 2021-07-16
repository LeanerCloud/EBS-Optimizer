package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EBSVolume extends ec2.Volume with a few useful things.
type EBSVolume struct {
	types.Volume
	api    ec2Conn
	region string
}

func (v *EBSVolume) process() error {

	log.Printf("Processing volume %s in %s\n", *v.VolumeId, v.region)
	vc := v.getCurrentConfiguration()
	nvc := v.newVolumeConfiguration()

	if vc.VolumeType == nvc.VolumeType {
		log.Printf("Volume configuration unchanged, skipping volume %s in %s\n", *v.VolumeId, v.region)
		return nil
	}
	log.Printf("Current volume configuration for %s in %s: %+v, new volume configuration: %+v \n", *v.VolumeId, v.region, vc, nvc)
	return v.modify(&nvc)
}

func (v *EBSVolume) modify(config *volumeConfig) error {

	v.backupConfiguration()

	if conf.DryRun {
		log.Printf("Dry-run: would modify volume %+v from %s to %s\n",
			*v.VolumeId, v.VolumeType, config.VolumeType)
		return nil
	}

	_, err := v.api.ec2.ModifyVolume(context.TODO(), &ec2.ModifyVolumeInput{
		VolumeId:   v.VolumeId,
		VolumeType: config.VolumeType,
	})

	if err != nil {
		log.Println("Couldn't modify volume", *v.VolumeId, err.Error())
		return err
	}

	return nil
}

func (v *EBSVolume) getCurrentConfiguration() *volumeConfig {
	vc := volumeConfig{
		VolumeType: v.VolumeType,
		IOPS:       *v.Iops,
		Throughput: v.getThroughput(),
		Region:     v.region,
		Size:       *v.Size,
	}
	debug.Printf("Current configuration for %s is %v", *v.VolumeId, vc)
	return &vc
}

func (v *EBSVolume) getInitialConfiguration() *volumeConfig {
	var vc volumeConfig
	for _, tag := range v.Tags {
		if *tag.Key == InitialConfigurationTag {
			val := *tag.Value
			err := json.Unmarshal([]byte(val), &vc)
			if err != nil {
				fmt.Println("error unmarshalling initial configuration tag:", err)
			}
			vc.Region, vc.Size = v.region, *v.Size
			return &vc
		}
	}
	return nil
}

func (v *EBSVolume) getThroughput() int32 {
	if v.Throughput == nil {
		return 0
	}
	return *v.Throughput
}

func (v *EBSVolume) newVolumeConfiguration() volumeConfig {
	var nvc volumeConfig
	// by default give the same configuration as original volume
	nvc.VolumeType = v.VolumeType

	if string(v.VolumeType) == "gp2" {
		nvc.VolumeType = "gp3" // always makes sense to convert to GP3 as per https://cloudwiry.com/ebs-gp3-vs-gp2-pricing-comparison/
		if *v.Size > 1000 && conf.GP3MatchGP2IOPS {
			nvc.IOPS = *v.Size * 3 // match GP2 IOPS for large volumes
		}
		if *v.Size > 170 && conf.GP3MatchGP2BurstThroughput {
			nvc.Throughput = 250 // match GP2 burstable throughput for smaller volumes
		}
	}

	// convert io1 to io2 in supported regions
	if string(v.VolumeType) == "io1" && io2Supports(v.region) {
		nvc.VolumeType = "io2"
	}

	// convert IO1 and IO2 volumes to gp3 if their PIOPS is smaller than the max GP3 PIOPS
	if (string(v.VolumeType) == "io1" || string(v.VolumeType) == "io2") && *v.Iops < 16000 {
		nvc.VolumeType = "gp3"
		nvc.IOPS = *v.Iops
		nvc.Throughput = *v.Throughput
	}

	return nvc
}

func (v *EBSVolume) backupConfiguration() {
	log.Println("Backing up configuration to tags")
	if !v.hasInitialConfigurationBackup() {
		log.Println("Missing initial configuration, backing it up")
		v.backupInitialConfiguration()
	}
	log.Println("Backing up current configuration")
	v.backupCurrentConfigurationAsPrevious()
}

func (v *EBSVolume) backupInitialConfiguration() {
	v.saveConfigurationToTag(InitialConfigurationTag)
}

func (v *EBSVolume) backupCurrentConfigurationAsPrevious() {
	v.saveConfigurationToTag(PreviousConfigurationTag)
}

func (v *EBSVolume) hasInitialConfigurationBackup() bool {
	for _, tag := range v.Tags {
		if *tag.Key == InitialConfigurationTag {
			return true
		}
	}
	return false
}

func (v *EBSVolume) saveConfigurationToTag(key string) {
	vc := v.getCurrentConfiguration()
	log.Printf("Current configuration for %s: %v", *v.VolumeId, vc)

	value := vc.toString()
	debug.Printf("Configuration %v converted to string: %s\n", vc, value)

	if conf.DryRun {
		log.Printf("Dry-run: would modify volume %s tag %s to %s\n",
			*v.VolumeId, key, value)
		return
	}

	v.api.ec2.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: []string{*v.VolumeId},
		Tags: []types.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	})
}

func (v *EBSVolume) calculateMonthlySavings() float64 {
	currentMonthlyCost := v.getCurrentConfiguration().calculateMonthlyPrice()
	debug.Printf("Current monthly cost for %s in %s: %f", *v.VolumeId, v.region, currentMonthlyCost)

	ic := v.getInitialConfiguration()
	if ic == nil {
		debug.Printf("Missing initial configuration for %s in %s", *v.VolumeId, v.region)
		return 0
	}

	initialMonthlyCost := ic.calculateMonthlyPrice()
	debug.Printf("Initial monthly cost for %s in %s: %f", *v.VolumeId, v.region, initialMonthlyCost)
	savings := initialMonthlyCost - currentMonthlyCost
	if savings > 0 {
		log.Printf("Monthly savings for %s in %s: %f", *v.VolumeId, v.region, savings)
	} else {
		// just in case we got a negative number after the user modified the volume manually
		savings = 0
	}
	return savings
}

func (v *EBSVolume) calculateHourlySavings() float64 {
	return v.calculateMonthlySavings() / 730
}
