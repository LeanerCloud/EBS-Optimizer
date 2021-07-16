package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/namsral/flag"
)

const (
	// DefaultGP2ConversionThreshold is the size under which GP3 is more performant than GP2 for both throughput and IOPS
	DefaultGP2ConversionThreshold = 170

	// InitialConfigurationTag is the name of the tag applied to the EBS volume that holds a backup of the initial configuration of the volume, in JSON format.
	InitialConfigurationTag = "ebs_optimizer_initial_configuration"
	// PreviousConfigurationTag is the name of the tag applied to the EBS volume that holds a backup of the previous configuration of the volume, in JSON format.
	PreviousConfigurationTag = "ebs_optimizer_previous_configuration"
)

// Config stores the global configuration
type Config struct {

	// Logging
	LogFile io.Writer
	LogFlag int

	// The regions where it should be running, given as a single CSV-string
	Regions string

	// The region where the Lambda function is deployed
	MainRegion string

	// This is only here for tests, where we want to be able to somehow mock
	// time.Sleep without actually sleeping. While testing it defaults to 0 (which won't sleep at all), in
	// real-world usage it's expected to be set to 1
	SleepMultiplier time.Duration

	// JSON file containing event data used for locally simulating execution from Lambda.
	EventFile string

	// Filter on volume tags
	FilterByTags string
	// Controls how are the tags used to filter the groups.
	// Available options: 'opt-in' and 'opt-out', default: 'opt-in'
	TagFilteringMode string

	// The ebs-optimizer version
	Version string

	// Final Recap String Array to show actions taken by ScheduleRun on ASGs
	FinalRecap map[string][]string

	// Controls whether GP3 volumes replacing GP2 should be configured with provisioned IOPS to match GP2 performance
	GP3MatchGP2IOPS bool

	// Controls whether GP3 volumes replacing GP2 should be configured with provisioned throughput to match GP2 performance
	GP3MatchGP2BurstThroughput bool

	// DryRun controls whether to run in dry-run mode (without applying any changes).
	DryRun bool
}

// ParseCommandlineFlags loads configuration from command line flags, environments variables, and config files.
func (c *Config) ParseCommandlineFlags() {

	// The use of FlagSet allows us to parse config multiple times, which is useful for unit tests.
	flagSet := flag.NewFlagSet("ebs-optimizer", flag.ExitOnError)

	var region string

	if r := os.Getenv("AWS_REGION"); r != "" {
		region = r
	} else {
		region = "us-east-1"
	}

	c.MainRegion = region
	c.SleepMultiplier = 1

	flagSet.StringVar(&conf.Regions, "regions", "",
		"\n\tRegions where it should be activated (separated by comma or whitespace, also supports globs).\n"+
			"\tBy default it runs on all regions.\n"+
			"\tExample: ./ebs-optimizer -regions 'eu-*,us-east-1'\n")

	flagSet.StringVar(&conf.TagFilteringMode, "tag_filtering_mode", "opt-out", "\n\tControls the behavior of the tag_filters option.\n"+
		"\tValid choices: opt-in | opt-out\n\tDefault value: 'opt-in'\n\tExample: ./ebs-optimizer --tag_filtering_mode opt-in\n")

	flagSet.StringVar(&conf.FilterByTags, "tag_filters", "", "\n\tSet of tags to filter the volumes on.\n"+
		"\tDefault if no value is set will be the equivalent of -tag_filters 'optimize=true'\n"+
		"\tIn case the tag_filtering_mode is set to opt-out, it defaults to 'optimize=false'\n"+
		"\tExample: ./ebs-optimizer --tag_filters 'optimize=true'\n")

	flagSet.BoolVar(&conf.GP3MatchGP2IOPS, "gp3_match_gp2_iops", false,
		"\n\tControls whether to configure GP3 volumes with provisioned IOPS to match "+
			"GP2 burst performance characteristics for the same volume size (ignored for volumes smaller than 1TB).\n"+
			"See https://cloudwiry.com/ebs-gp3-vs-gp2-pricing-comparison/ for a more detailed explanation"+
			"\tExample: ./ebs-optimizer --gp3_match_gp2_iops true\n")

	flagSet.BoolVar(&conf.GP3MatchGP2BurstThroughput, "gp3_match_gp2_throughput", false,
		"\n\tControls whether to configure GP3 volumes with provisioned throughput to match "+
			"GP2 burst performance characteristics for the same volume size (ignored for volumes smaller than 170GB)\n"+
			"See https://cloudwiry.com/ebs-gp3-vs-gp2-pricing-comparison/ for a more detailed explanation\n"+
			"\tExample: ./ebs-optimizer --gp3_match_gp2_throughput true\n")

	flagSet.BoolVar(&conf.DryRun, "dry_run", false, "Run in dry-run mode, just show what it would do, without applying any changes.")

	printVersion := flagSet.Bool("version", false, "Print version number and exit.\n")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error parsing config: %s\n", err.Error())
	}

	if *printVersion {
		fmt.Println("ebs-optimizer build:", conf.Version)
		os.Exit(0)
	}

	c.FinalRecap = make(map[string][]string)
}
