![Hello World](data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEYAAAAUCAAAAAAVAxSkAAABrUlEQVQ4y+3TPUvDQBgH8OdDOGa+oUMgk2MpdHIIgpSUiqC0OKirgxYX8QVFRQRpBRF8KShqLbgIYkUEteCgFVuqUEVxEIkvJFhae3m8S2KbSkcFBw9yHP88+eXucgH8kQZ/jSm4VDaIy9RKCpKac9NKgU4uEJNwhHhK3qvPBVO8rxRWmFXPF+NSM1KVMbwriAMwhDgVcrxeMZm85GR0PhvGJAAmyozJsbsxgNEir4iEjIK0SYqGd8sOR3rJAGN2BCEkOxhxMhpd8Mk0CXtZacxi1hr20mI/rzgnxayoidevcGuHXTC/q6QuYSMt1jC+gBIiMg12v2vb5NlklChiWnhmFZpwvxDGzuUzV8kOg+N8UUvNBp64vy9q3UN7gDXhwWLY2nMC3zRDibfsY7wjEkY79CdMZhrxSqqzxf4ZRPXwzWJirMicDa5KwiPeARygHXKNMQHEy3rMopDR20XNZGbJzUtrwDC/KshlLDWyqdmhxZzCsdYmf2fWZPoxCEDyfIvdtNQH0PRkH6Q51g8rFO3Qzxh2LbItcDCOpmuOsV7ntNaERe3v/lP/zO8yn4N+yNPrekmPAAAAAElFTkSuQmCC)
# ACME webhook mijn.host
This is a Cert Manager webhook for [mijn.host](https://mijn.host) DNS.
This webhook is tested with cert-manager: v1.17.2

Please note that i'm not an expert in Go so don't expect perfect and clean code.
I made this because nothing existed yet for mijn.host DNS.

## Requirements
* Cert-manager
* mijn.host [API key](https://mijn.host/cp/account/api/#)

## Installation
This webhook can be installed with Helm.

```bash
helm -n cert-manager install cert-manager-webhook-mijnhost ./deploy/cert-manager-webhook-mijnhost
```
Please change the namespace and serviceAccountNames in values.yml when using a different deployment name or namespace.

## Issuer
1. Encode the mijn.host API key in Base64 and create a secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mijnhost-api-key
  namespace: cert-manager
type: Opaque
data:
  applicationSecret: APIKEY_BASE64
```

2. Create issuer
I used a ClusterIssuer in this example

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: mijnhost-issuer
spec:
  acme:
    email: jouw@email.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: mijnhost-issuer-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.mijnhost
          solverName: mijnhost
          config:
            applicationSecretRef:
              name: mijnhost-api-key
              key: applicationSecret
```

3. Create certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
spec:
  secretName: example-com-tls
  issuerRef:
    name: mijnhost-issuer
    kind: ClusterIssuer
  dnsNames:
  - "example.com"
  - "testz0r.example.com"
```

# Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

You can run the test suite with:

```bash
$ TEST_ZONE_NAME=example.com. make test
```

