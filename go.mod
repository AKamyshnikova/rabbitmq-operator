module github.com/toha10/rabbitmq-operator

require (
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.19.13
	k8s.io/apimachinery v0.19.13
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	sigs.k8s.io/controller-runtime v0.6.5
	github.com/go-logr/logr v0.4.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.19.13 // Required by prometheus-operator
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
)

go 1.16
