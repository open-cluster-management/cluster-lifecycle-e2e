# E2E failed analysis
## Quota limit in aws/azure/gcp
The e2e failed because of that there are not enough resources to provision an ocp cluster in aws/azure/gcp.
It needs CICD team to cleanup the resources or add the related resources in aws/gcp/azure.
**Please contact CICD team in slack channel: #forum-acm-devops**
 
The error messages should be like the following:
- VPC: VpcLimitExceeded: The maximum number of VPCs has been reached."
- Quota 'EXTERNAL_NETWORK_LB_FORWARDING_RULES' exceeded. Limit: 36.0 in region us-east1."
- Error: Code=\"PublicIPCountLimitReached\" Message=\"Cannot create more than 60 public IP addresses for this subscription in this region.\" Details=[]"
- compute.googleapis.com/cpus is not available in us-east1 because the required number of resources (24) is more than remaining quota of 12"
- ...
 
## Cloud provider(aws/gcp/azure) bug or ocp installer bug
Sometimes the e2e will fail because the cloud provider is not stable.

**If the error only occurs once in one day, it should be cloud provider not stable, and it's difficult to reproduce. You should rerun the e2e.**
**If the error always occurs for one provider(aws/gcp/azure), you need to ask the issue in channel #forum-installer.**

The error should be like:
- Cluster initialization failed because one or more operators are not functioning properly
- Failed calling webhook \"clusterdeploymentvalidators.admission.hive.openshift.io\": the server is currently unable to handle the request"
- This is a bug in the provider, which should be reported in the provider's own"

## Known issues for clc
### klusterlet CRD can not be deleted
There is a known issue for ACM 2.3, the klusterlet crd may not be deleted when detaching a cluster.
**Please ignore the error and rerun the e2e.**
