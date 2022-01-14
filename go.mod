module github.com/stolostron/cluster-lifecycle-e2e

go 1.16

require (
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.1
	github.com/open-cluster-management/library-e2e-go v0.0.0-20210219164826-04b5828bcc17
	github.com/open-cluster-management/library-go v0.0.0-20200828173847-299c21e6c3fc
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/text v0.3.7 // indirect
	k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2

)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/coreos/etcd => go.etcd.io/etcd v3.3.22+incompatible
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v0.19.0
	github.com/open-cluster-management/library-e2e-go => github.com/stolostron/library-e2e-go v0.0.0-20210219164826-04b5828bcc17
	github.com/open-cluster-management/library-go => github.com/stolostron/library-go v0.0.0-20200828173847-299c21e6c3fc
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
