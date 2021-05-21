#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.

echo "Initiating tests..."
echo "Tests start $TEST_GROUP at "$(date)

if [[ $TEST_GROUP == "import" ]]; then
    ginkgo -v -focus="import" -trace -debug import_cluster/import_cluster.test -- -v=3
elif [[ $TEST_GROUP == "provision-all" ]]; then
    ginkgo -v -focus="create" --nodes=3 -trace -debug create_cluster/create_cluster.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -cloud-providers=aws,azure,gcp
elif [[ $TEST_GROUP == "destroy" ]]; then
    ginkgo -v -focus="detach|destroy" --nodes=4 -trace -debug detach_destroy/detach_destroy.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -cloud-providers=aws,azure,gcp
elif [[ $TEST_GROUP == "metrics" ]]; then
    ginkgo -v -focus="metrics" -trace -debug metrics/metrics.test -- -v=3
elif [[ $TEST_GROUP == "create-baremetal" ]]; then
    ginkgo -v -focus="create" -trace -debug create_cluster_bm/create_cluster_bm.test -- -v=3 -cloud-providers=baremetal
elif [[ $TEST_GROUP == "destroy-baremetal" ]]; then
    ginkgo -v -focus="destroy" -trace -debug destroy_bm/destroy_bm.test -- -v=3 -cloud-providers=baremetal
fi

echo "Tests end $TEST_GROUP at "$(date)