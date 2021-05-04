#!/usr/bin/env bash

minikube start -p nri-postgresql-testInt --driver=docker --insecure-registry
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install db-postgresql -f build/k8s/postgresql.yml bitnami/postgresql

eval $(minikube docker-env)
nri-postgresql
kubectl run nri-postgresqlv1 --image=nri-postgresql:0.0.1 --image-pull-policy=Never

kubectl create deployment nri-postgresql --image=newrelic/nri-postgresql:local --imagePullPolicy=NEVER
kubectl get deployments
kubectl get pods
kubectl get events
kubectl expose deployment nri-postgresql --type=LoadBalancer --port=80
minikube service nri-postgresql

minikube stop -p nri-postgresql-testInt
minikube delete -p nri-postgresql-testInt