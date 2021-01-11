#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.

echo "Initiating tests..."

if [[ $TEST_GROUP == "import" ]]; then
    ginkgo -v -focus="import" --reportFile=$REPORT_FILE_IMPORT -trace -debug e2e-test.test -- -v=3
elif [[ $TEST_GROUP == "provision-all" ]]; then
    ginkgo -v -focus="create" -p -trace --debug e2e-test.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -report-file=$REPORT_FILE_CREATE
elif [[ $TEST_GROUP == "destroy" ]]; then
    ginkgo -v -focus="detach|destroy" -p -trace -debug e2e-test.test -- -v=3 -owner="ginkgo-$TRAVIS_BUILD_ID" -report-file=$REPORT_FILE_DESTROY
elif [[ $TEST_GROUP == "metrics" ]]; then
    ginkgo -v -focus="metrics" -skip="" --reportFile=$REPORT_FILE_METRICS -trace -debug e2e-test.test -- -v=3
elif [[ $TEST_GROUP == "baremetal" ]]; then
    ginkgo -v  -focus="baremetal" -skip="" -trace --debug e2e-test.test -- -v=3 
fi