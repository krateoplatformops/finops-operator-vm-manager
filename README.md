# finops-operator-vm-manager
This repository is part of a wider exporting architecture for the FinOps Cost and Usage Specification (FOCUS) in Kratep. This component is tasked with applying an optimization to an Azure resource, according to the description given in a Custom Resource (CR).

## Configuration
To apply an optimization, see the "config-sample.yaml" file.
The deployment of the operator needs a secret for the repository, called `registry-credentials` in the namespace `finops`.

## Installation
### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Installation with HELM
```sh
$ helm repo add krateo https://charts.krateo.io
$ helm repo update krateo
$ helm install finops-operator-vm-manager krateo/finops-operator-vm-manager
```

## Bearer-token for Azure
In order to invoke Azure API, the exporter needs to be authenticated first. In the current implementation, it utilizes the Azure REST API, which require the bearer-token for authentication. For each target Azure subscription, an application needs to be registered and assigned with the Cost Management Reader role.

Once that is completed, run the following command to obtain the bearer-token (1h validity):
```
curl -X POST -d 'grant_type=client_credentials&client_id=<CLIENT_ID>&client_secret=<CLIENT_SECRET>&resource=https%3A%2F%2Fmanagement.azure.com%2F' https://login.microsoftonline.com/<TENANT_ID>/oauth2/token
```