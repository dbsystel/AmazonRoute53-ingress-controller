:warning: The project has been archived and is no longer maintained!

# Amazon Route53 Ingress Controller

This Controller watches for new *ingress resources* and if they define the specified annotations as `true`, it will create an Amazon Route53 record set.

Ingress example can be found [here](ingress-resource-examples).

## Annotations

`ingress.net/route53` with values: `"true"` or `"false"`

`ingress.net/load-balancer-name: "load-balancer-name"`:  Specify load balancer name. Created Amazon Route53 record will have an alias pointing to provided loadbalancer. As of now ELB and ALB are supported.

**Note**

Mentioned `"true"` values can be also specified with: `"1", "t", "T", "true", "TRUE", "True"`

Mentioned `"false"` values can be also specified with: `"0", "f", "F", "false", "FALSE", "False"`

## Usage
```
--run-outside-cluster # Uses ~/.kube/config rather than in cluster configuration
--log-level # desired log level, one of: [debug, info, warn, error]
--log-format # desired log format, one of: [json, logfmt]
--allowlist-prefix # comma sperated list with Amazon Route53 record name prefixes, which has to be matched, before update/delete Amazon Route53 record sets 
--allowlist-suffix # comma sperated list with Amazon Route53 record name suffixes, which has to be matched, before update/delete Amazon Route53 record sets 
--delete-alias # if true, recordset type alias will be deleted before other recordset type being created.
--delete-cname # if true, recordset type cname will be deleted before other recordset type being created.
--dns-type # DNS Record Type(alias / cname), default cname
```

Example:
`./bin/AmazonRoute53-ingress-controller --run-outside-cluster --log-level=info --allowlist-suffix=example.local,test.local --allowlist-prefix=app-`

For example, with provided allowlist the following Amazon Route53 records could be created/updated/deleted:
- test.example.local
- example-test.local
- app-domain.local

For example, with provided allowlist the following Amazon Route53 records could *not* be created/updated/deleted:
- app.domain.local
- apps-test.local

## Access
The Amazon Route53 Ingress Controller needs to know, in which AWS region you are operating it. Please set your AWS region as environment variable, e.g.:
- `export AWS_REGION=eu-central-1`

For authentification with the Amazon Route53 API you can either use IAM roles, attached to your nodes, or you have to provide two additional environment variables:
- `export AWS_ACCESS_KEY_ID=XXX`
- `export AWS_SECRET_ACCESS_KEY=XXX`
 
Or run `aws configure`, if you have installed the `aws-cli`.

If you want to deploy the controller via Helm, all three variables can be provided in `values.yaml`, see example installation at our [Helm directory](helm) within this repo.

## Development
### Build
```
CGO_ENABLED=0 go build -v -i -o ./bin/AmazonRoute53-ingress-controller ./cmd # on Linux
GOOS=linux CGO_ENABLED=0 go build -v -i -o ./bin/AmazonRoute53-ingress-controller ./cmd # on macOS/Windows
```

### Run outside kubernetes
```
export AWS_REGION=eu-central-1 #make sure AWS_REGION is set
./bin/AmazonRoute53-ingress-controller --run-outside-cluster --log-level=debug
```

## Deployment
Our preferred way to install AmazonRoute53-ingress-controller is [Helm](https://helm.sh/). See example installation at our [Helm directory](helm) within this repo.
