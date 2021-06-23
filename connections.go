package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type ec2Conn struct {
	session *session.Session
	ec2     ec2iface.EC2API
	region  string
}

func (c *ec2Conn) setSession(region string) {
	c.session = session.Must(
		session.NewSession(&aws.Config{Region: aws.String(region)}))
}

func (c *ec2Conn) connect(region, mainRegion string) {

	log.Println("Creating service connections in", region)

	if c.session == nil {
		c.setSession(region)
	}

	ec2Conn := make(chan *ec2.EC2)

	go func() { ec2Conn <- ec2.New(c.session) }()

	c.ec2, c.region = <-ec2Conn, region

	log.Println("Created service connections in", region)
}
