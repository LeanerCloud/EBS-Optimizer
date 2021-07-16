package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type ec2Conn struct {
	config *aws.Config
	ec2    *ec2.Client
	region string
}

func (c *ec2Conn) connect(region, mainRegion string) {

	debug.Println("Creating service connections in", region)

	if c.config == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
		if err != nil {
			log.Fatalf("Could not create EC2 service connections")
		}
		c.config = &cfg
	}

	ec2Conn := make(chan *ec2.Client)

	go func() { ec2Conn <- ec2.NewFromConfig(*c.config) }()

	c.ec2, c.region = <-ec2Conn, region

	debug.Println("Created service connections in", region)
}
