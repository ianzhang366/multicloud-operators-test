#!/bin/bash

cd ..
operator-sdk generate k8s
operator-sdk generate crds
operator-sdk build quay.io/stevekimibm/multicloud-operators-test
docker login -u="stevekimibm" -p="y5TryF1llhPRf6LfSEIQwBYOD4Q8Bbp/RJI7a47/P32RUCYXkhPmemqG9GrYQchL" quay.io
docker push quay.io/stevekimibm/multicloud-operators-test
