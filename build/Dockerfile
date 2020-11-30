FROM openshift/origin-release:golang-1.13 as builder

ARG REMOTE_SOURCE
ARG REMOTE_SOURCE_DIR
ARG GITHUB_TOKEN
ARG COMPONENT_VERSION

ENV USE_VENDORIZED_BUILD_HARNESS=true
ENV COMPONENT_NAME=cluster-lifecycle-e2e
ENV COMPONENT_VERSION="${COMPONENT_VERSION}"

# Install ChromeDriver in build image
# ENV CHROMEDRIVER_VERSION "80.0.3987.16"
# RUN mkdir -p $REMOTE_SOURCE_DIR/chromedriver
# ENV CHROMEDRIVER_VERSION "81.0.4044.69"
# RUN wget --no-check-certificate -N https://chromedriver.storage.googleapis.com/${CHROMEDRIVER_VERSION}/chromedriver_linux64.zip -P $REMOTE_SOURCE_DIR/chromedriver

# RUN mkdir -p $REMOTE_SOURCE_DIR/app/build/bin
# RUN unzip $REMOTE_SOURCE_DIR/chromedriver/chromedriver_linux64.zip -d $REMOTE_SOURCE_DIR/app/build/bin

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
ENV GINKGO_NODES "1"
ENV GINKGO_FLAGS=""
ENV GINKGO_FOCUS=""
ENV GINKGO_SKIP=""

ENV E2E_BIN=/usr/local/bin/e2e-test.test \
    USER_UID=1001 \
    USER_NAME=cluster-lifecycle

# install operator binary
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/e2e-test.test ${E2E_BIN}
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin /usr/local/bin

# Copy resources which contains the options.yaml
COPY --from=builder $REMOTE_SOURCE_DIR/app/e2e-test/resources /test/resources

RUN  /usr/local/bin/user_setup

VOLUME /results

# execute compiled ginkgo tests
ENTRYPOINT ["/usr/local/bin/entrypoint-e2e"]

USER ${USER_UID}