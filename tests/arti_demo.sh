#!/bin/bash

echo "user: aug_e4_intranet_jenkins_writer"
echo "password:"
kubectl get $(kubectl get secrets -n kube-system -l user=aug_e4_intranet_jenkins_writer -o name) -o template --template='{{.data.password}}' | base64 -d
echo "###### Docker Login ####"
docker login aug-e4-docker-scratch-intranet.registry.saas.cagip.gca
docker push aug-e4-docker-scratch-intranet.registry.saas.cagip.gca/busybox:latest
kubectl -n aug-e4-development apply -f tests/deploy-busybox.yml