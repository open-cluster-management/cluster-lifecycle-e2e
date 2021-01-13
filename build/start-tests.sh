#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.

echo "Initiating tests..."

if [[ $TEST_GROUP == "import" ]]; then
    ginkgo -v -focus="import" -skip="" --reportFile=$REPORT_FILE_IMPORT -trace -debug e2e-test.test -- -v=3
elif [[ $TEST_GROUP == "provision-all" ]]; then
    ginkgo -v -focus="create" -skip="" -p -trace --debug e2e-test.test -- -v=3
elif [[ $TEST_GROUP == "destroy" ]]; then
    ginkgo -v -focus="detach" -skip="" --reportFile=$REPORT_FILE_DETACH -trace -debug e2e-test.test -- -v=3
elif [[ $TEST_GROUP == "metrics" ]]; then
    ginkgo -v -focus="metrics" -skip="" --reportFile=$REPORT_FILE_DETACH -trace -debug e2e-test.test -- -v=3
fi