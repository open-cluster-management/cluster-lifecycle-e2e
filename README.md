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
   - monitor deletiion of clusterdeployment and cluster-namespace
2. Import already exiisting cluster 
   - monitor cluster and add-ons to be iin ready state
   - detach imported cluster 
   - monitor deletion of cluster-namespace
3. Check if local-cluster is imported and in ready state 

## Running E2E

1. clone this repo:

```
$ git clone git@github.com:open-cluster-management/cluster-lifecycle-e2e.git
```

2. copy `resources/options.yaml.template` to `resources/options.yaml`, and update values specific to your environment:

```
$ cp resources/options.yaml.template resources/options.yaml
```

3. run testing:

```
$ export KUBECONFIG=~/.kube/config
$ ginkgo -v -p -stream -- -options=resources/options.yaml -v=3
```

## Running with Docker

1. clone this repo:

```
$ git clone git@github.com:open-cluster-management/cluster-lifecycle-e2e.git
```

2. copy `resources/options.yaml.template` to `resources/options.yaml`, and update values specific to your environment:

```
$ cp resources/options.yaml.template resources/options.yaml
```

3. oc login to your hub cluster where you want to run these tests - and make sure that remains the current-context in kubeconfig:

```
$ kubectl config current-context
open-cluster-management/api-demo-dev02-red-chesterfield-com:6443/kube:admin
```

4. run `make build`. This will create a docker image:

```
$ make build
```

5. run the following command to get docker image ID, we will use this in the next step:

```
$ docker_image_id=`docker images | grep cluster-lifecycle-e2e | sed -n '1p' | awk '{print $3}'`
```

6. run testing:

```
$ docker run -v ~/.kube/:/opt/.kube -v $(pwd)/results:/results -v $(pwd)/e2e-test/resources:/resources $docker_image_id
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
