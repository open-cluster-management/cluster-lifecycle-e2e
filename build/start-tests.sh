#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.

echo "Initiating tests..."

if [[ $TEST_GROUP == "import" ]]; then
    ginkgo -v -focus="import" --reportFile=$REPORT_FILE_IMPORT -trace -debug import_cluster/import_cluster.test -- -v=3
elif [[ $TEST_GROUP == "provision-all" ]]; then
    ginkgo -v -focus="create" --nodes=3 -trace -debug create_cluster/create_cluster.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -report-file=$REPORT_FILE_CREATE -cloud-providers=aws,azure,gcp
elif [[ $TEST_GROUP == "destroy" ]]; then
    ginkgo -v -focus="detach|destroy" -p -trace -debug detach_destroy/detach_destroy.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -report-file=$REPORT_FILE_DESTROY -cloud-providers=aws,azure,gcp
elif [[ $TEST_GROUP == "metrics" ]]; then
    ginkgo -v -focus="metrics" --reportFile=$REPORT_FILE_METRICS -trace -debug metrics/metrics.test -- -v=3
elif [[ $TEST_GROUP == "create-baremetal" ]]; then
    ginkgo -v -focus="create" --reportFile=$REPORT_FILE_BM -trace -debug create_cluster_bm/create_cluster_bm.test -- -v=3 -cloud-providers=baremetal
elif [[ $TEST_GROUP == "destroy-baremetal" ]]; then
    ginkgo -v -focus="destroy" --reportFile=$REPORT_FILE_BM -trace -debug destroy_bm/destroy_bm.test -- -v=3 -cloud-providers=baremetal
fi