FROM openshift/origin-release:golang-1.13 as builder

ARG REMOTE_SOURCE
ARG REMOTE_SOURCE_DIR
ARG GITHUB_TOKEN
ARG COMPONENT_VERSION

ENV USE_VENDORIZED_BUILD_HARNESS=true
ENV COMPONENT_NAME=cluster-lifecycle-e2e
ENV COMPONENT_VERSION="${COMPONENT_VERSION}"


ENV GOPATH /go
# ENV CGO_ENABLED=0

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
ENV OPTIONS "/resources/options.yaml"
ENV REPORT_FILE "/results/results.xml"
ENV GINKGO_DEFAULT_FLAGS "-slowSpecThreshold=120 -timeout 7200s"
ENV GINKGO_NODES "3"
ENV GINKGO_FLAGS=""
ENV GINKGO_FOCUS=""
ENV GINKGO_SKIP=""

# run as root
USER root

# ENV E2E_BIN=/usr/local/bin/e2e-test.test \
#     USER_UID=root \
#     USER_NAME=cluster-lifecycle

# install ginkgo into built image
COPY --from=builder /go/bin/ /usr/local/bin

# Copy resources which contains the options.yaml
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/resources /test/resources

# install operator binary
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/e2e-test.test /test/resources
#COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin /usr/local/bin


VOLUME /results
#WORKDIR "/test/resources"
# execute compiled ginkgo tests
#ENTRYPOINT ["/usr/local/bin/entrypoint-e2e"]

CMD ["/bin/bash", "-c", "ginkgo --v --focus=${GINKGO_FOCUS} --tags=e2e --skip=${GINKGO_SKIP} --p --reportFile=${REPORT_FILE} /test/resources/e2e-test.test -- -v=3"]
#USER ${USER_UID}
#ginkgo -p -tags e2e e2e-test -- -v=5