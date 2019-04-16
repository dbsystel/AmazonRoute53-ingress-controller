package controller

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dbsystel/AmazonRoute53-ingress-controller/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"k8s.io/api/extensions/v1beta1"
)

// define struct
type Controller struct {
	logger          log.Logger
	whitelistPrefix string
	whitelistSuffix string
}

// create new object from type Controller and return object pointer
func New(logger log.Logger, whitelistPrefix string, whitelistSuffix string) *Controller {
	controller := &Controller{}
	controller.logger = logger
	controller.whitelistPrefix = whitelistPrefix
	controller.whitelistSuffix = whitelistSuffix
	return controller
}

// do something when an ingress resource is beeing created
func (c *Controller) Create(obj interface{}) {
	level.Debug(c.logger).Log("msg", "Called function: Create")
	ingressObj := obj.(*v1beta1.Ingress)

	r53, _ := ingressObj.Annotations["ingress.net/route53"]

	isR53, _ := strconv.ParseBool(r53)

	if isR53 {
		level.Info(c.logger).Log("msg", "Creation of an ingress resource detected", "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)

		c.createRecordSet(ingressObj)
	}
}

// do something when an ingress resource is beeing updated
func (c *Controller) Update(oldobj interface{}, newobj interface{}) {
	newIngressObj := newobj.(*v1beta1.Ingress)
	oldIngressObj := oldobj.(*v1beta1.Ingress)

	level.Debug(c.logger).Log("msg", "Called function: Update")

	if c.noDifference(oldIngressObj, newIngressObj) {
		level.Debug(c.logger).Log("msg", "Skipping automatically updated ingress", "ingressName", newIngressObj.Name, "ingressNamespace", newIngressObj.Namespace)
		return
	}

	r53, _ := newIngressObj.Annotations["ingress.net/route53"]
	isR53, _ := strconv.ParseBool(r53)

	if isR53 {
		level.Info(c.logger).Log("msg", "Update of an ingress resource detected", "ingressName", newIngressObj.Name, "ingressNamespace", newIngressObj.Namespace)

		c.deleteRecordSet(oldIngressObj)

		c.createRecordSet(newIngressObj)
	}
}

// do something when an ingress resource is beeing deleted
func (c *Controller) Delete(obj interface{}) {
	level.Debug(c.logger).Log("msg", "Called function: Delete")
	ingressObj := obj.(*v1beta1.Ingress)

	r53, _ := ingressObj.Annotations["ingress.net/route53"]

	isR53, _ := strconv.ParseBool(r53)

	if isR53 {
		level.Info(c.logger).Log("msg", "Deletion of an ingress resource detected", "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)

		c.deleteRecordSet(ingressObj)
	}
}

func (c *Controller) searchHostedZoneId(host string) string {

	hostedZoneId, err := aws.GetHostedZone(host, c.logger)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeInvalidInput:
				level.Error(c.logger).Log("err", route53.ErrCodeInvalidInput, "msg", aerr.Error())
			case route53.ErrCodeNoSuchDelegationSet:
				level.Error(c.logger).Log("err", route53.ErrCodeNoSuchDelegationSet, "msg", aerr.Error())
			case route53.ErrCodeDelegationSetNotReusable:
				level.Error(c.logger).Log("err", route53.ErrCodeDelegationSetNotReusable, "msg", aerr.Error())
			default:
				level.Error(c.logger).Log("msg", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			level.Error(c.logger).Log("msg", err.Error())
		}
	}

	return hostedZoneId
}

// retrun dnsName and hostedZoneNameID for give load-balancer-name
func (c *Controller) getLoadBalancerAttributes(loadBalancername string) (string, string) {

	dnsName, hostedZoneNameID := aws.GetELBAttributes(loadBalancername, c.logger)
	if dnsName == "" {
		dnsName, hostedZoneNameID = aws.GetALBAttributes(loadBalancername, c.logger)
	}
	return dnsName, hostedZoneNameID
}

// are two ingress resources same?
func (c *Controller) noDifference(newIngressObj *v1beta1.Ingress, oldIngressObj *v1beta1.Ingress) bool {
	if len(newIngressObj.Spec.Rules) != len(oldIngressObj.Spec.Rules) {
		newIngressObjContent, _ := json.Marshal(newIngressObj.Spec.Rules)
		oldIngressObjContent, _ := json.Marshal(oldIngressObj.Spec.Rules)
		level.Debug(c.logger).Log(
			"msg", "length of ingressObj spec rules are different",
			"newIngressObjSpecRulesLength", len(newIngressObj.Spec.Rules),
			"oldIngressObjSpecRulesLength", len(oldIngressObj.Spec.Rules),
			"newIngressObjSpecRulesContent", string(newIngressObjContent),
			"oldIngressObjSpecRulesContent", string(oldIngressObjContent),
			)
		return false
	}
	for i, ingressRule := range newIngressObj.Spec.Rules {
		if ingressRule.Host != oldIngressObj.Spec.Rules[i].Host {
			level.Debug(c.logger).Log(
				"msg", "ingressObj spec rules host names are different",
				"newIngressObjHostName", ingressRule.Host,
				"oldIngressObjHostName", oldIngressObj.Spec.Rules[i].Host,
			)
			return false
		}
	}

	if newIngressObj.Annotations["ingress.net/load-balancer-name"] != oldIngressObj.Annotations["ingress.net/load-balancer-name"] {
		level.Debug(c.logger).Log(
			"msg", "ingressObj annotations load-balancer-name are different",
			"newIngressObjAnnotation", newIngressObj.Annotations["ingress.net/load-balancer-name"],
			"oldIngressObjAnnotation", oldIngressObj.Annotations["ingress.net/load-balancer-name"],
			)
		return false
	}
	return true
}

// check if gice host is in whitelist
func (c *Controller) isInWhitelist(host string) (inWhitelist bool) {
	if c.whitelistPrefix != "" {
		prefixes := strings.Split(c.whitelistPrefix, ",")
		for _, prefix := range prefixes {
			if prefix == "" {
				continue
			}
			inWhitelist = strings.HasPrefix(host, prefix)
			if inWhitelist {
				break
			}

		}
	}
	if c.whitelistSuffix != "" {
		suffixes := strings.Split(c.whitelistSuffix, ",")
		for _, suffix := range suffixes {
			if suffix == "" {
				continue
			}
			if inWhitelist {
				break
			}
			inWhitelist = strings.HasSuffix(host, suffix)
		}
	}
	return inWhitelist
}

// delete Amazon Route53 recordset
func (c *Controller) deleteRecordSet(ingressObj *v1beta1.Ingress) {
	loadBalancerName, _ := ingressObj.Annotations["ingress.net/load-balancer-name"]

	for _, ingressRule := range ingressObj.Spec.Rules {
		level.Info(c.logger).Log("msg", "Deleting Route53 record set", "hostName", ingressRule.Host, "ingressName", ingressRule.Host, "ingressNamespace", ingressObj.Namespace)
		if c.isInWhitelist(ingressRule.Host) {
			hostedZoneId := c.searchHostedZoneId(ingressRule.Host)
			level.Debug(c.logger).Log("msg", "Found Hosted Zone ID: ", "hostedzoneid", hostedZoneId)

			aliasName, aliasHostedZoneId := c.getLoadBalancerAttributes(loadBalancerName)
			level.Debug(c.logger).Log("aliasName: ", aliasName, "aliasHostedZoneId: ", aliasHostedZoneId)
			result, err := aws.ChangeRecordSet("DELETE", aliasName, aliasHostedZoneId, ingressRule.Host, hostedZoneId)

			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case route53.ErrCodeNoSuchHostedZone:
						level.Error(c.logger).Log("err", route53.ErrCodeNoSuchHostedZone, "msg", aerr.Error())
					case route53.ErrCodeNoSuchHealthCheck:
						level.Error(c.logger).Log("err", route53.ErrCodeNoSuchHealthCheck, "msg", aerr.Error())
					case route53.ErrCodeInvalidChangeBatch:
						re := regexp.MustCompile("but it already exists")
						if re.Match([]byte(aerr.Error())) {
							level.Info(c.logger).Log("err", route53.ErrCodeInvalidChangeBatch, "msg", aerr.Error())
						} else {
							level.Error(c.logger).Log("err", route53.ErrCodeInvalidChangeBatch, "msg", aerr.Error())
						}
					case route53.ErrCodeInvalidInput:
						level.Error(c.logger).Log("err", route53.ErrCodeInvalidInput, "msg", aerr.Error())
					case route53.ErrCodePriorRequestNotComplete:
						level.Error(c.logger).Log("err", route53.ErrCodeInvalidInput, "msg", aerr.Error())
					default:
						level.Error(c.logger).Log("msg", aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					level.Error(c.logger).Log("msg", err.Error())
				}
			} else {
				level.Info(c.logger).Log("msg", result, "hostName", ingressRule.Host, "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)
			}

		} else {
			level.Info(c.logger).Log("msg", "Provided host "+ingressRule.Host+" is not in whitelist. Skipping deletion!", "hostName", ingressRule.Host, "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)
		}
	}

}

// create Amazon Route53 recordset
func (c *Controller) createRecordSet(ingressObj *v1beta1.Ingress) {
	loadBalancerName, _ := ingressObj.Annotations["ingress.net/load-balancer-name"]

	for _, ingressRule := range ingressObj.Spec.Rules {
		level.Info(c.logger).Log("msg", "Creating/Updating Route53 record set", "hostName", ingressRule.Host, "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)
		if c.isInWhitelist(ingressRule.Host) {

			hostedZoneId := c.searchHostedZoneId(ingressRule.Host)
			level.Debug(c.logger).Log("msg", "Found Hosted Zone ID: ", "hostedzoneid", hostedZoneId)

			aliasName, aliasHostedZoneId := c.getLoadBalancerAttributes(loadBalancerName)
			level.Debug(c.logger).Log("aliasName: ", aliasName, "aliasHostedZoneId: ", aliasHostedZoneId)

			result, err := aws.ChangeRecordSet("UPSERT", aliasName, aliasHostedZoneId, ingressRule.Host, hostedZoneId)

			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case route53.ErrCodeNoSuchHostedZone:
						level.Error(c.logger).Log("err", route53.ErrCodeNoSuchHostedZone, "msg", aerr.Error())
					case route53.ErrCodeNoSuchHealthCheck:
						level.Error(c.logger).Log("err", route53.ErrCodeNoSuchHealthCheck, "msg", aerr.Error())
					case route53.ErrCodeInvalidChangeBatch:
						re := regexp.MustCompile("but it already exists")
						if re.Match([]byte(aerr.Error())) {
							level.Info(c.logger).Log("err", route53.ErrCodeInvalidChangeBatch, "msg", aerr.Error())
						} else {
							level.Error(c.logger).Log("err", route53.ErrCodeInvalidChangeBatch, "msg", aerr.Error())
						}
					case route53.ErrCodeInvalidInput:
						level.Error(c.logger).Log("err", route53.ErrCodeInvalidInput, "msg", aerr.Error())
					case route53.ErrCodePriorRequestNotComplete:
						level.Error(c.logger).Log("err", route53.ErrCodeInvalidInput, "msg", aerr.Error())
					default:
						level.Error(c.logger).Log("msg", aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					level.Error(c.logger).Log("msg", err.Error())
				}
			} else {
				level.Info(c.logger).Log("msg", result, "hostName", ingressRule.Host, "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)
			}
		} else {
			level.Info(c.logger).Log("msg", "Provided host "+ingressRule.Host+" is not in whitelist. Skipping creation/updating!", "hostName", ingressRule.Host, "ingressName", ingressObj.Name, "ingressNamespace", ingressObj.Namespace)
		}
	}
}
