# DTM

## Usage

To install the dtm chart:

    helm install --create-namespace -n dtm-system dtm .

To install the dtm chart:

    helm upgrade -n dtm-system dtm .

To uninstall the chart:

    helm delete -n dtm-system dtm

## Parameters

### Configuration parameters

| Key           | Description                                        | Value |
| ------------- | -------------------------------------------------- | ----- |
| configuration | DTM configuration. Specify content for config.yaml | ""    |

### Ingress parameters

| Key                          | Description                                                                     | Value             |
| ---------------------------- | ------------------------------------------------------------------------------- | ----------------- |
| ingress.enabled              | Enable ingress record generation for DTM                                        | false             |
| ingress.className            | IngressClass that will be be used to implement the Ingress (Kubernetes 1.18+)   | "nginx"           |
| ingress.annotations          | To enable certificate autogeneration, place here your cert-manager annotations. | {}                |
| ingress.hosts.host           | Default host for the ingress record.                                            | "your-domain.com" |
| ingress.hosts.paths.path     | Default path for the ingress record                                             | "/"               |
| ingress.hosts.paths.pathType | Ingress path type                                                               | "Prefix"          |
| ingress.tls                  | Enable TLS configuration for the host defined at ingress.hostname parameter     | []                |
