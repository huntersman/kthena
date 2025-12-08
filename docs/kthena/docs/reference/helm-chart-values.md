# Helm Chart Values

This document provides a comprehensive reference for all configurable values in the Kthena Helm chart.

## Overview

The Kthena Helm chart consists of two main subcharts: `workload` and `networking`, along with global settings. Each subchart configures different components of the Kthena system.

## Global Values

Global values apply across all subcharts and control cluster‑wide settings.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `global.certManagementMode` | `string` | `"auto"` | Certificate management mode. Valid values: `auto` (self‑signed), `cert‑manager` (use cert‑manager), `manual` (provide your own CA bundle). |
| `global.webhook.caBundle` | `string` | `""` | Base64‑encoded CA bundle for webhook server certificates. Required only when `certManagementMode` is `"manual"`. |

## Workload Subchart

The workload subchart deploys the Kthena Controller Manager, which manages ModelServings, ModelBoosters, and autoscaling policies.

### Controller Manager

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `workload.enabled` | `boolean` | `true` | Enable or disable the workload subchart. |
| `workload.controllerManager.image.repository` | `string` | `"ghcr.io/volcano-sh/kthena-controller-manager"` | Container image repository. |
| `workload.controllerManager.image.tag` | `string` | `"latest"` | Container image tag (usually set by CI). |
| `workload.controllerManager.image.pullPolicy` | `string` | `"IfNotPresent"` | Image pull policy. |
| `workload.controllerManager.image.args` | `list` | `["--v=2"]` | Command‑line arguments passed to the controller manager. |
| `workload.controllerManager.replicas` | `integer` | `1` | Number of controller manager replicas. |
| `workload.controllerManager.resource.limits.cpu` | `string` | `"500m"` | CPU limit. |
| `workload.controllerManager.resource.limits.memory` | `string` | `"512Mi"` | Memory limit. |
| `workload.controllerManager.resource.requests.cpu` | `string` | `"100m"` | CPU request. |
| `workload.controllerManager.resource.requests.memory` | `string` | `"128Mi"` | Memory request. |
| `workload.controllerManager.webhook.enabled` | `boolean` | `true` | Enable the admission webhook. |
| `workload.controllerManager.webhook.tls.certSecretName` | `string` | `"kthena-controller-manager-webhook-certs"` | Name of the secret storing auto‑generated certificates. |
| `workload.controllerManager.webhook.tls.serviceName` | `string` | `"kthena-controller-manager-webhook"` | Service name used for certificate DNS names. |
| `workload.controllerManager.downloaderImage.repository` | `string` | `"ghcr.io/volcano-sh/downloader"` | Downloader container image repository. |
| `workload.controllerManager.downloaderImage.tag` | `string` | `"latest"` | Downloader image tag. |
| `workload.controllerManager.runtimeImage.repository` | `string` | `"ghcr.io/volcano-sh/runtime"` | Runtime container image repository. |
| `workload.controllerManager.runtimeImage.tag` | `string` | `"latest"` | Runtime image tag. |
| `workload.controllerManager.downloader.accessKey` | `string` | `""` | Access key for the downloader (if required). |
| `workload.controllerManager.downloader.secretKey` | `string` | `""` | Secret key for the downloader (if required). |

## Networking Subchart

The networking subchart deploys the Kthena Router and its associated webhook, which handles request routing, rate limiting, and traffic policies.

### Kthena Router

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `networking.enabled` | `boolean` | `true` | Enable or disable the networking subchart. |
| `networking.kthenaRouter.enabled` | `boolean` | `true` | Enable the Kthena Router component. |
| `networking.kthenaRouter.replicas` | `integer` | `1` | Number of router replicas. |
| `networking.kthenaRouter.port` | `integer` | `8080` | HTTP port the router listens on. |
| `networking.kthenaRouter.image.repository` | `string` | `"ghcr.io/volcano-sh/kthena-router"` | Router container image repository. |
| `networking.kthenaRouter.image.tag` | `string` | `"latest"` | Router image tag (usually set by CI). |
| `networking.kthenaRouter.image.pullPolicy` | `string` | `"IfNotPresent"` | Image pull policy. |
| `networking.kthenaRouter.resource.limits.cpu` | `string` | `"500m"` | CPU limit. |
| `networking.kthenaRouter.resource.limits.memory` | `string` | `"512Mi"` | Memory limit. |
| `networking.kthenaRouter.resource.requests.cpu` | `string` | `"100m"` | CPU request. |
| `networking.kthenaRouter.resource.requests.memory` | `string` | `"128Mi"` | Memory request. |
| `networking.kthenaRouter.tls.enabled` | `boolean` | `false` | Enable TLS for the router. |
| `networking.kthenaRouter.tls.dnsName` | `string` | `""` | DNS name for the TLS certificate (required if TLS enabled). |
| `networking.kthenaRouter.tls.secretName` | `string` | `"kthena-router-tls"` | Name of the secret storing the TLS certificate and key. |
| `networking.kthenaRouter.webhook.enabled` | `boolean` | `true` | Enable the router webhook. |
| `networking.kthenaRouter.webhook.port` | `integer` | `8443` | Port the webhook listens on. |
| `networking.kthenaRouter.webhook.servicePort` | `integer` | `443` | Service port for the webhook. |
| `networking.kthenaRouter.webhook.tls.certFile` | `string` | `"/etc/tls/tls.crt"` | Path to the TLS certificate file inside the pod. |
| `networking.kthenaRouter.webhook.tls.keyFile` | `string` | `"/etc/tls/tls.key"` | Path to the TLS key file inside the pod. |
| `networking.kthenaRouter.webhook.tls.secretName` | `string` | `"kthena-router-webhook-certs"` | Name of the secret storing webhook certificates. |
| `networking.kthenaRouter.webhook.tls.serviceName` | `string` | `"kthena-router-webhook"` | Service name used for certificate DNS names. |
| `networking.kthenaRouter.fairness.enabled` | `boolean` | `false` | Enable fairness scheduling for request prioritization. |
| `networking.kthenaRouter.fairness.windowSize` | `string` | `"1h"` | Sliding window duration for token usage tracking (e.g., `1m`, `5m`, `1h`). |
| `networking.kthenaRouter.fairness.inputTokenWeight` | `float` | `1.0` | Weight multiplier for input tokens in priority calculation. |
| `networking.kthenaRouter.fairness.outputTokenWeight` | `float` | `2.0` | Weight multiplier for output tokens in priority calculation. |
| `networking.kthenaRouter.accessLog.enabled` | `boolean` | `true` | Enable access logging. |
| `networking.kthenaRouter.accessLog.format` | `string` | `"text"` | Log format: `"json"` or `"text"`. |
| `networking.kthenaRouter.accessLog.output` | `string` | `"stdout"` | Where to write logs: `"stdout"`, `"stderr"`, or a file path. |

### Router Webhook

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `networking.webhook.enabled` | `boolean` | `true` | Enable the router webhook component. |
| `networking.webhook.replicas` | `integer` | `1` | Number of webhook replicas. |
| `networking.webhook.image.repository` | `string` | `"ghcr.io/volcano-sh/kthena-router-webhook"` | Webhook container image repository. |
| `networking.webhook.image.tag` | `string` | `"latest"` | Webhook image tag (usually set by CI). |
| `networking.webhook.image.pullPolicy` | `string` | `"IfNotPresent"` | Image pull policy. |
| `networking.webhook.image.args` | `list` | `["--port=8443"]` | Command‑line arguments passed to the webhook. |
| `networking.webhook.resource.limits.cpu` | `string` | `"100m"` | CPU limit. |
| `networking.webhook.resource.limits.memory` | `string` | `"128Mi"` | Memory limit. |
| `networking.webhook.resource.requests.cpu` | `string` | `"100m"` | CPU request. |
| `networking.webhook.resource.requests.memory` | `string` | `"128Mi"` | Memory request. |

## Default Values

The complete default `values.yaml` file is shown below:

```yaml
workload:
  # enabled is a flag to enable or disable the workload subchart. Default is true.
  enabled: true
  controllerManager:
    image:
      repository: ghcr.io/volcano-sh/kthena-controller-manager
      tag: latest
      pullPolicy: IfNotPresent
    downloaderImage:
      repository: ghcr.io/volcano-sh/downloader
      tag: latest
    runtimeImage:
      repository: ghcr.io/volcano-sh/runtime
      tag: latest
    webhook:
      enabled: true
      tls:
        certSecretName: kthena-controller-manager-webhook-certs
        serviceName: kthena-controller-manager-webhook

networking:
  # enabled is a flag to enable or disable the networking subchart. Default is true.
  enabled: true
  kthenaRouter:
    enabled: true
    port: 8080
    image:
      repository: ghcr.io/volcano-sh/kthena-router
      tag: latest
      pullPolicy: IfNotPresent
    tls:
      enabled: false
      # The DNS name to use for the certificate.
      dnsName: "your-domain.com"
      # The name of the secret to store the certificate and key.
      secretName: "kthena-router-tls"
    webhook:
      enabled: true
      port: 8443
      servicePort: 443
      tls:
        certFile: /etc/tls/tls.crt
        keyFile: /etc/tls/tls.key
        secretName: kthena-router-webhook-certs
    # fairness configuration for request scheduling
    fairness:
      # enabled controls whether fairness scheduling is active
      enabled: false
      # windowSize is the sliding window duration for token usage tracking
      windowSize: "1h"
      # inputTokenWeight is the weight multiplier for input tokens
      inputTokenWeight: 1.0
      # outputTokenWeight is the weight multiplier for output tokens
      outputTokenWeight: 2.0

global:
  # Certificate Management Mode
  # Three mutually exclusive options for managing TLS certificates:
  #   - auto: Webhook servers generate self-signed certificates automatically (default)
  #   - cert-manager: Use cert-manager to generate and manage certificates (requires cert-manager installation)
  #   - manual: Provide your own certificates via caBundle
  certManagementMode: auto
  webhook:
    # caBundle is the base64-encoded CA bundle for webhook server certificates.
    # This is ONLY required when certManagementMode is set to "manual".
    # You can generate it with: cat /path/to/your/ca.crt | base64 | tr -d '\n'
    caBundle: ""
```

## Notes

- Values marked as “usually set by CI” are automatically updated during the release process; manual changes are not required.
- For detailed information about each component, refer to the corresponding architecture and user guide documents.
- Always review the [values.yaml](https://github.com/volcano-sh/kthena/blob/main/charts/kthena/values.yaml) file in the repository for the latest defaults and available options.