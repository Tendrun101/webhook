module webhook

go 1.16

require (
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/klog/v2 v2.40.1
)

replace k8s.io/apimachinery v0.23.4 => k8s.io/apimachinery v0.20.16-rc.0
