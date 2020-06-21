# Helm Chart for Route53 Ingress Controller
### Installing the Chart

To install the chart with the release name `amazonroute53-ingress-controller` in namespace `ops`:

```console
$ helm upgrade amazonroute53-ingress-controller charts/AmazonRoute53-ingress-controller --namespace ops --install
```
The command deploys AmazonRoute53-ingress-controller on the Kubernetes cluster with the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

### Uninstalling the Chart

To uninstall/delete the `amazonroute53-ingress-controller ` deployment:

```console
$ helm delete amazonroute53-ingress-controller  --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
The following table lists the configurable parameters of the AmazonRoute53-ingress-controller chart.

| Parameter | Description | Default |
| --- | --- | --- |
| `image.repository`                      | Image | `dockerregistry/devops/amazonroute53-ingress-controller` |
| `image.tag`                             | Image tag  | `1.2.0` |
| `logLevel`                              | Desired log level, one of: [debug, info, warn, error]  | `info` |
| `logFormat`                             | Desired log format, one of: [json, logfmt]  | `json` |
| `accessKey`                             | If you want to set own AWS Access Key ID just alter "false" to "YOURCUSTOMID"  | `false` |
| `secretKey`                             | If you want to set own AWS Secret Access Key just alter "false" to "YOURCUSTOMKEY | `false` |
| `allowlistPrefix`                       | For safety reasons only Amazon Route53 recods will be created/updated/deleted if they match with the allowlist. At least one allowlist (prefix or suffix or both) should be always provided (as csv).  | `awesome` |
| `allowlistSuffix`                       | For safety reasons only Amazon Route53 recods will be created/updated/deleted if they match with the allowlist. At least one allowlist (prefix or suffix or both) should be always provided (as csv).  | `mytestdomain.com,mytestdomain.org` |
| `replicaCount`                          | Desired number of pods | `1` |
| `resources`                             | Pod resource requests & limits | `{"limits": { "cpu": "100m", "memory": "100Mi" }, "requests": {"cpu": "100m", "memory": "100Mi" }}` |

## Installing the Chart

```console
$ helm upgrade amazonroute53-ingress-controller  charts/AmazonRoute53-ingress-controller/ --namespace ops --install
```

## Uninstalling the Chart

To uninstall `amazonroute53-ingress-controller`:

```console
$ helm delete --purge amazonroute53-ingress-controller 
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

> **Tip**: List all releases using `helm list` or start clean with `helm delete --purge amazonroute53-ingress-controller`

