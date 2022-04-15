# Eclipse Che Helm Charts

- [Charts](#charts)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)


## Charts

Helm charts to deploy [Eclipse Che](https://www.eclipse.org/che/)

### Prerequisites

* Minimal Kubernetes version is 1.19
* Minimal Helm version is 3.2.2
* [Cert manager](https://cert-manager.io/docs/installation/) installed

### Installation
Install the Helm Charts for Eclipse Che Operator

```
helm install che \
  --set ingress.domain=<KUBERNETES_INGRESS_DOMAIN> \
  --set ingress.oauth.oAuthSecret=<OAUTH_SECRET> \
  --set ingress.oauth.oAuthClientName=<OAUTH_CLIENT_NAME> \
  --set ingress.oauth.identityProviderURL=<IDENTITY_PROVIDER_URL> .
```
