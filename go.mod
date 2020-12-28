module github.com/jenkins-x-apps/jx-app-jacoco

require (
	github.com/GeertJohan/fgt v0.0.0-20160120143236-262f7b11eec0 // indirect
	github.com/antham/gommit v2.2.0+incompatible // indirect
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/cenkalti/backoff v2.0.0+incompatible
	github.com/dlclark/regexp2 v1.1.6 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/golang/lint v0.0.0-20181217174547-8f45f776aaf1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.2
	github.com/jenkins-x/jx v1.3.1069
	github.com/magiconair/properties v1.8.0
	github.com/pkg/errors v0.8.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/viper v1.3.1 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	k8s.io/api v0.0.0-20190126160303-ccdd560a045f
	k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e
	k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/metrics v0.0.0-20190205053707-d984453de47b // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190130003954-e5e029740eb8 // indirect
)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190318120422-253758214767

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7

replace k8s.io/client-go => k8s.io/client-go v2.0.0-alpha.0.0.20190115164855-701b91367003+incompatible

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
