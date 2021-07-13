package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EBSVolume struct {
	*ec2.Volume
	api    ec2Conn
	region string
}

func (v *EBSVolume) process() error {

	log.Println("Processing volume ", v)
	vc := v.getCurrentConfiguration()
	// vc.calculatePrice()

	log.Printf("Current volume configuration: %+v\n", vc)

	nvc := v.newVolumeConfiguration()

	log.Printf("New volume configuration: %+v\n", nvc)

	if vc == nvc {
		log.Println("Volume configuration unchanged, skipping volume", *v.VolumeId)
		return nil
	}
	return v.modify(&nvc)
}

func (v *EBSVolume) modify(config *volumeConfig) error {

	v.backupConfiguration()

	if conf.DryRun {
		log.Printf("Dry-run: would modify volume %+v from %s to %s\n",
			*v.VolumeId, *v.VolumeType, config.VolumeType)
		return nil
	}

	_, err := v.api.ec2.ModifyVolume(&ec2.ModifyVolumeInput{
		VolumeId:   v.VolumeId,
		VolumeType: aws.String(config.VolumeType),
	})

	if err != nil {
		log.Println("Couldn't modify volume", *v.VolumeId, err.Error())
		return err
	}

	return nil
}

func (v *EBSVolume) getCurrentConfiguration() volumeConfig {
	vc := volumeConfig{
		VolumeType: *v.VolumeType,
		IOPS:       *v.Iops,
		Throughput: v.getThroughput(),
	}
	log.Printf("Current configuration for %s is %v", *v.VolumeId, vc)
	return vc
}

func (v *EBSVolume) getThroughput() int64 {
	if v.Throughput == nil {
		return 0
	} else {
		return *v.Throughput
	}
}

func (v *EBSVolume) newVolumeConfiguration() volumeConfig {
	var nvc volumeConfig
	// by default give the same configuration as original volume
	nvc.VolumeType = *v.VolumeType

	if *v.VolumeType == "gp2" {
		nvc.VolumeType = "gp3" // always makes sense to convert to GP3 as per https://cloudwiry.com/ebs-gp3-vs-gp2-pricing-comparison/
		if *v.Size > 1000 && conf.GP3MatchGP2IOPS {
			nvc.IOPS = *v.Size * 3 // match GP2 IOPS for large volumes
		}
		if *v.Size > 170 && conf.GP3MatchGP2BurstThroughput {
			nvc.Throughput = 250 // match GP2 burstable throughput for smaller volumes
		}
	}

	// convert io1 to io2 in supported regions
	if *v.VolumeType == "io1" && io2Supports(v.region) {
		nvc.VolumeType = "io2"
	}

	// convert IO1 and IO2 volumes to gp3 if their PIOPS is smaller than the max GP3 PIOPS
	if (*v.VolumeType == "io1" || *v.VolumeType == "io2") && *v.Iops < 16000 {
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
	log.Printf("Configuration %s converted to string: %v", vc, value)

	if conf.DryRun {
		log.Printf("Dry-run: would modify volume %s tag %s to %s\n",
			*v.VolumeId, key, value)
		return
	}

	v.api.ec2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{v.VolumeId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	})
}
