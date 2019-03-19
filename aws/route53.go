package aws

import (
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// define struct
type HostedZone struct {
	name string
	id   string
}

// modify Amazon Route53 recordset with given state (upsert/delete)
func ChangeRecordSet(state, aliasName, aliasHostedZoneId, name, hostedZoneId string) (string, error) {
	sess := session.Must(session.NewSession())
	svc := route53.New(sess)

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(state),
					ResourceRecordSet: &route53.ResourceRecordSet{
						AliasTarget: &route53.AliasTarget{
							DNSName:              aws.String(aliasName),
							EvaluateTargetHealth: aws.Bool(true),
							HostedZoneId:         aws.String(aliasHostedZoneId),
						},
						Name: aws.String(name),
						Type: aws.String("A"),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// return Hosted Zone ID for provided host
func GetHostedZone(host string, logger log.Logger) (string, error) {
	host = host + "."
	level.Debug(logger).Log("msg", "Searching Hosted Zone ID for provided host ", "host", host)
	sess := session.Must(session.NewSession())
	svc := route53.New(sess)

	input := &route53.ListHostedZonesInput{}
	output, err := svc.ListHostedZones(input)

	if err != nil {
		return "", err
	}

	reg := regexp.MustCompile("^/hostedzone/")
	for _, hostedZone := range output.HostedZones {

		if strings.HasSuffix(host, "."+*hostedZone.Name) {

			return reg.ReplaceAllString(*hostedZone.Id, ""), nil
		}
	}

	return "", errors.New("Hosted Zone ID for provided string: " + host + " not found!")
}
