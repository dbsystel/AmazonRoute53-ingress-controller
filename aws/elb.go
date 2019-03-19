package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// return elb attributes for provided load-balancer-name
func GetELBAttributes(loadBalancername string, logger log.Logger) (string, string) {
	sess := session.Must(session.NewSession())
	svc := elb.New(sess)

	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(loadBalancername),
		},
	}
	output, err := svc.DescribeLoadBalancers(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elb.ErrCodeAccessPointNotFoundException:
				level.Debug(logger).Log("info", elb.ErrCodeAccessPointNotFoundException, "msg", aerr.Error())
			case elb.ErrCodeDependencyThrottleException:
				level.Error(logger).Log("err", elb.ErrCodeDependencyThrottleException, "msg", aerr.Error())
			default:
				level.Error(logger).Log("msg", err.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			level.Error(logger).Log("msg", err.Error())
		}
	}

	for _, loadBalancerDescription := range output.LoadBalancerDescriptions {

		return *loadBalancerDescription.DNSName, *loadBalancerDescription.CanonicalHostedZoneNameID

	}
	return "", ""
}

// return alb attributes for provided load-balancer-name
func GetALBAttributes(loadBalancername string, logger log.Logger) (string, string) {
	sess := session.Must(session.NewSession())
	svc := elbv2.New(sess)

	input := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{
			aws.String(loadBalancername),
		},
	}
	output, err := svc.DescribeLoadBalancers(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				level.Error(logger).Log("err", elbv2.ErrCodeLoadBalancerNotFoundException, "msg", aerr.Error())
			default:
				level.Error(logger).Log("msg", err.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			level.Error(logger).Log("msg", err.Error())
		}
	}

	for _, loadBalancer := range output.LoadBalancers {

		return *loadBalancer.DNSName, *loadBalancer.CanonicalHostedZoneId

	}
	return "", ""
}
