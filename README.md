# cluster-lifecycle-e2e
An e2e test lib for cluster lifecycle

This is a container which will be called from:

1. Canary Tests

This will be called after ACM is installed.

The tests in this container will:
1. Provision aws, gcp, azure cluster  
   - Create necessary resources to provision cluster
   - monitor cluster and add-ons to be in ready state
   - destroy cluster
   - monitor deletion of clusterdeployment and cluster-namespace
2. Import already existing cluster
   - monitor cluster and add-ons to be in ready state
   - detach imported cluster
   - monitor deletion of cluster-namespace
3. Check if local-cluster is imported and in ready state

## Running E2E

1. clone this repo:

```
$ git clone git@github.com:open-cluster-management/cluster-lifecycle-e2e.git
```

2. copy `e2e-test/resources/options.yaml.template` to `e2e-test/resources/options.yaml`, and update values specific to your environment:

```
$ cp e2e-test/resources/options.yaml.template e2e-test/resources/options.yaml
```

3. run testing:

```
$ export KUBECONFIG=~/.kube/config
$ ginkgo -v -p -stream -- -options=e2e-test/resources/options.yaml -v=3
```

## Running with Docker

1. clone this repo:

```
$ git clone git@github.com:open-cluster-management/cluster-lifecycle-e2e.git
$ cd cluster-lifecycle-e2e
```

2. copy `e2e-test/resources/options_template.yaml` to `e2e-test/resources/options.yaml`, and update values specific to your environment:

```
$ cp e2e-test/resources/options_template.yaml e2e-test/resources/options.yaml
```

3. oc login to your hub cluster where you want to run these tests - and make sure that remains the current-context in kubeconfig:

```
$ kubectl config current-context
open-cluster-management/api-demo-dev02-red-chesterfield-com:6443/kube:admin
```

4. Set env variables `GITHUB_USER` and `GITHUB_TOKEN` and Run `make deps`. This will download necessary dependencies

```
$ make deps
```

5. Run `make build`. This will create a docker image:

```
$ make build
```

6. run the following command to get docker image ID, we will use this in the next step:

```
$ EXPORT docker_image_id=`docker images | grep cluster-lifecycle-e2e | sed -n '1p' | awk '{print $3}'`
```

7. run testing:

TEST_GROUP values can be
- import -> to import an existing cluster
- provision-all -> to provision aws, gcp, azure clusters in parallel
- destroy -> to deatch an existing imported cluster
- metrics -> to test the clusterlifecycle metrics from prometheus
- baremetal -> to provision baremetal cluster

For import test, save kubeconfig of cluster to be imported in path `$(pwd)/e2e-test/resources/import/kubeconfig`

```
$ docker run -v ~/.kube/config:/opt/.kube/config -v $(pwd)/e2e-test/resources/import/kubeconfig:/opt/.kube/import-kubeconfig -v $(pwd)/results:/results -v $(pwd)/e2e-test/resources:/resources -v $(pwd)/e2e-test/resources/options.yaml:/resources/options.yaml  --env TEST_GROUP="import" $docker_image_id
```

In Canary environment, this is the container that will be run - and all the volumes etc will passed on while starting the docker container using a helper script.

## Contributing to E2E

### Options.yaml

The values in the options.yaml are optional values read in by E2E. If you do not set an option, the test case that depends on the option should skip the test. The sample values in the option.yaml.template should provide enough context for you fill in with the appropriate values. Further, in the section below, each test should document their test with some detail.

### Focus Labels

* The `--focus` and `--skip` are ginkgo directives that allow you to choose what tests to run, by providing a REGEX express to match. Examples of using the focus:

  * `ginkgo --focus="g0"`
  * `ginkgo --focus="import"`

* To run with verbose ginkgo logging pass the `--v`
* To run with klog verbosity, pass the `--focus="g0" -- -v=3` where 3 is the log level: 1-3
