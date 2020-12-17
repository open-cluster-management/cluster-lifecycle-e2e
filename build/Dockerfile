FROM openshift/origin-release:golang-1.13 as builder

ARG REMOTE_SOURCE
ARG REMOTE_SOURCE_DIR
ARG GITHUB_TOKEN
ARG COMPONENT_VERSION

ENV USE_VENDORIZED_BUILD_HARNESS=true
ENV COMPONENT_NAME=cluster-lifecycle-e2e
ENV COMPONENT_VERSION="${COMPONENT_VERSION}"


ENV GOPATH /go

ENV PATH usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/go/bin:.

COPY $REMOTE_SOURCE $REMOTE_SOURCE_DIR/app/
WORKDIR $REMOTE_SOURCE_DIR/app
# RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

# compile go tests in build image
RUN GOFLAGS="" go get -u github.com/onsi/ginkgo/ginkgo && GOFLAGS="" ginkgo build -tags e2e e2e-test 

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ARG VCS_REF
ARG VCS_URL
ARG IMAGE_NAME
ARG IMAGE_DESCRIPTION
ARG ARCH_TYPE
ARG REMOTE_SOURCE_DIR

RUN microdnf update

# expose env vars for runtime
ENV KUBECONFIG "/opt/.kube/config"
ENV IMPORT_KUBECONFIG "/opt/.kube/import-kubeconfig"
ENV OPTIONS "/resources/options.yaml"
ENV REPORT_FILE_IMPORT "/results/results-import.xml"
#ENV REPORT_FILE_CREATE "/results/results-provision.xml"
ENV REPORT_FILE_DETACH "/results/results-detach.xml"
ENV GINKGO_DEFAULT_FLAGS "-slowSpecThreshold=120 -timeout 7200s"
ENV GINKGO_NODES "11"
ENV GINKGO_FLAGS=""
ENV GINKGO_FOCUS=""
ENV GINKGO_SKIP=""

# run as root
USER root

# install ginkgo into built image
COPY --from=builder /go/bin/ /usr/local/bin

# Copy resources which contains the options.yaml
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/resources /test/resources

# install operator binary
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/e2e-test.test /test
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/start-tests.sh /test/start-tests.sh 
VOLUME /results
WORKDIR "/test"

#CMD ["/bin/bash", "-c", "ginkgo --v --focus=${GINKGO_FOCUS} --skip=${GINKGO_SKIP} --p --trace --debug e2e-test.test -- -v=3"]

RUN ["chmod", "+x", "start-tests.sh"]

ENTRYPOINT ["./start-tests.sh"]