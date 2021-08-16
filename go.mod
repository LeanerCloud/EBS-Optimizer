module github.com/cristim/ebs-optimizer

go 1.16

require (
	github.com/aws/aws-lambda-go v1.24.0
	github.com/aws/aws-sdk-go-v2 v1.8.0
	github.com/aws/aws-sdk-go-v2/config v1.5.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.12.0
	github.com/aws/aws-sdk-go-v2/service/marketplacemetering v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/pricing v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.9.0 // indirect
	github.com/mattn/goveralls v0.0.9
	github.com/namsral/flag v1.7.4-pre
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/tools v0.1.1
)
