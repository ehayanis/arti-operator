#!/bin/bash

 kubectl delete secrets -l creator=artifactory-operator && \
 kubectl delete ns aug-e4-development && \
 kubectl delete project aug-e4-development && \
 kubectl delete -f deployment/ -n kube-system && \
 kubectl create ns aug-e4-development && \
 kubectl apply -f deployment/ && \
 kubectl apply -f tests/project-test-resource.yaml