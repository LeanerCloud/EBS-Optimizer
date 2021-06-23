package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EBSVolume struct {
	*ec2.Volume
	api    ec2Conn
	region string
}

type volumeConfig struct {
	volumeType   string
	iops         int64
	throughput   int64
	monthlyPrice float64
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

	if conf.DryRun {
		log.Printf("Dry-run: would modify volume %+v from %s to %s\n",
			*v.VolumeId, *v.VolumeType, config.volumeType)
		return nil
	}
	_, err := v.api.ec2.ModifyVolume(&ec2.ModifyVolumeInput{
		VolumeId:   v.VolumeId,
		VolumeType: aws.String(config.volumeType),
	})

	if err != nil {
		log.Println("Couldn't modify volume", *v.VolumeId, err.Error())
		return err
	}

	return nil
}

func (v *EBSVolume) getCurrentConfiguration() volumeConfig {
	return volumeConfig{
		volumeType: *v.VolumeType,
		// iops:       *v.Iops,
		//throughput: v.getThroughput(),
	}
}

func (v *EBSVolume) newVolumeConfiguration() volumeConfig {
	var nvc volumeConfig
	// by default give the same configuration as original volume
	nvc.volumeType = *v.VolumeType

	if *v.VolumeType == "gp2" {
		nvc.volumeType = "gp3" // always makes sense to convert to GP3 as per https://cloudwiry.com/ebs-gp3-vs-gp2-pricing-comparison/
		if *v.Size > 1000 && conf.GP3MatchGP2IOPS {
			nvc.iops = *v.Size * 3 // match GP2 IOPS for large volumes
		}
		if *v.Size > 170 && conf.GP3MatchGP2BurstThroughput {
			nvc.throughput = 250 // match GP2 burstable throughput for smaller volumes
		}
	}

	// convert io1 to io2 in supported regions
	if *v.VolumeType == "io1" && io2Supports(v.region) {
		nvc.volumeType = "io2"
	}

	// convert IO1 and IO2 volumes to gp3 if their PIOPS is smaller than the max GP3 PIOPS
	if (*v.VolumeType == "io1" || *v.VolumeType == "io2") && *v.Iops < 16000 {
		nvc.volumeType = "gp3"
		nvc.iops = *v.Iops
		nvc.throughput = *v.Throughput
	}

	return nvc
}

func (vc *volumeConfig) calculatePrice() float64 {
	return 0
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
