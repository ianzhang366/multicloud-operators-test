#!/bin/bash

# Install script for building, pushing, and installing updated controller for a KinD cluster

set +x 
cd ..
echo "===> Reinstall KinD Cluster"
kind delete cluster
kind create cluster

echo "===> Connect to KinD cluster"
KIND_CONFIG=$(kind get kubeconfig-path)
export KUBECONFIG=$KIND_CONFIG

echo "===> Create ConfigMap"
kubectl create configmap kubeconfigs --from-file=$HOME/.kube/example-configmap

echo "===> Deploy operator and CRD"
kubectl apply -f deploy/crds/app.ibm.com_apptests_crd.yaml
kubectl apply -f deploy
kubectl apply -f deploy/crs/guestbook_cr.yaml

echo "===> Finished update!"

