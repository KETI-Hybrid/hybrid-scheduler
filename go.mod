module hybrid-scheduler

go 1.21.1

replace (
	k8s.io/api => k8s.io/api v0.22.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.0
	k8s.io/apiserver => k8s.io/apiserver v0.22.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.0
	k8s.io/client-go => k8s.io/client-go v0.22.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.0
	k8s.io/code-generator => k8s.io/code-generator v0.22.0
	k8s.io/component-base => k8s.io/component-base v0.22.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.0
	k8s.io/cri-api => k8s.io/cri-api v0.22.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.0
	k8s.io/kubectl => k8s.io/kubectl v0.22.0
	k8s.io/kubelet => k8s.io/kubelet v0.22.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.0
	k8s.io/metrics => k8s.io/metrics v0.22.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.0
)

require (
	github.com/KETI-Hybrid/keti-controller v0.0.0-20231004021214-46b5948d5b74
	github.com/julienschmidt/httprouter v1.3.0
	github.com/spf13/cobra v1.7.0
	google.golang.org/grpc v1.56.0
	google.golang.org/protobuf v1.30.0
	k8s.io/api v0.28.1
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.100.1
	k8s.io/kube-scheduler v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/term v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
	sigs.k8s.io/controller-runtime v0.14.1 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
