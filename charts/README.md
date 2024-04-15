# DTM charts

## Usage

Install the dtm chart:

```bash
helm install --create-namespace -n dtm-system dtm ./charts
```

Upgrade the dtm chart:

```bash
helm upgrade -n dtm-system dtm ./charts
```

Uninstall the chart:

```bash
helm delete -n dtm-system dtm
```

## Parameters

### Configuration parameters

| Key             | Description                                                                                                                           | Value |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------|-------|
| `configuration` | DTM configuration. Specify content for `config.yaml`, ref: [sample config](https://github.com/dtm-labs/dtm/blob/main/conf.sample.yml) | `""`  |



### Autoscaling Parameters

| Name                                            | Description                               | Value   |
|-------------------------------------------------|-------------------------------------------|---------|
| `autoscaling.enabled`                           | Enable Horizontal POD autoscaling for DTM | `false` |
| `autoscaling.minReplicas`                       | Minimum number of DTM replicas            | `1`     |
| `autoscaling.maxReplicas`                       | Maximum number of DTM replicas            | `10`    |
| `autoscaling.targetCPUUtilizationPercentage`    | Target CPU utilization percentage         | `80`    |
| `autoscaling.targetMemoryUtilizationPercentage` | Target Memory utilization percentage      | `80`    |

### Ingress parameters

| Key                            | Description                                                                   | Value               |
|--------------------------------|-------------------------------------------------------------------------------|---------------------|
| `ingress.enabled`              | Enable ingress record generation for DTM                                      | `false`             |
| `ingress.className`            | IngressClass that will be used to implement the Ingress (Kubernetes 1.18+)    | `"nginx"`           |
| `ingress.annotations`          | To enable certificate auto generation, place here your cert-manager annotations. | `{}`                |
| `ingress.hosts.host`           | Default host for the ingress record.                                          | `"your-domain.com"` |
| `ingress.hosts.paths.path`     | Default path for the ingress record                                           | `"/"`               |
| `ingress.hosts.paths.pathType` | Ingress path type                                                             | `"Prefix"`          |
| `ingress.tls`                  | Enable TLS configuration for the host defined at ingress.hostname parameter   | `[]`                |
