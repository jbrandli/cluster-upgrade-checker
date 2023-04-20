# cluster-upgrade-checker

## Setup

`go mod tidy` to install dependencies.

## usage

The program will use your default cluster and the credential setup at `~/.kube/config` You can use `kubectx` to switch the default cluster

`go run .` will create three files, one for HA deployments which should not have downtime, one for deploys with potential downtime. and one for deployments which could block node scaling operations.



