#!/usr/bin/env bash

minikube start -p nri-postgresql-test

helm repo add bitnami https://charts.bitnami.com/bitnami
helm install my-release \
  --set image.tag=13 bitnami/postgresql


minikube stop -p nri-postgresql-test
minikube delete -p nri-postgresql-test