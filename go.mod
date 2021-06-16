module github.com/open-cluster-management/cluster-lifecycle-e2e

go 1.16

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/open-cluster-management/applier v0.0.0-20210422205113-6c10f923726b
	//github.com/open-cluster-management/endpoint-operator v1.0.1-2020-05-28-18-29-00.0.20201123150704-b95ad181d88a // indirect
	github.com/open-cluster-management/library-e2e-go v0.0.0-20210219164826-04b5828bcc17
	github.com/open-cluster-management/library-go v0.0.0-20200828173847-299c21e6c3fc
	k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2

)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/coreos/etcd => go.etcd.io/etcd v3.3.22+incompatible
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v0.19.0
	//github.com/open-cluster-management/library-e2e-go => /Users/ljawale/go/src/github.com/open-cluster-management/library-e2e-go
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
