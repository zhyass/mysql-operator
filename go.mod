module github.com/zhyass/mysql-operator

go 1.15

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-ini/ini v1.62.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/imdario/mergo v0.3.11
	github.com/kr/text v0.2.0 // indirect
	github.com/nxadm/tail v1.4.8
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/presslabs/controller-util v0.3.0-alpha.2
	github.com/spf13/cobra v1.1.1
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)
