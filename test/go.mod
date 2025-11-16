module github.com/DataDog/helm-charts/test

<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
go 1.24.9
=======
go 1.24.0
>>>>>>> e760684d (changed the go version)
=======
go 1.24.9
>>>>>>> 7844aff2 (updated the dependecies)
=======
go 1.24.0

toolchain go1.24.4
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
=======
go 1.24.9
>>>>>>> 0b53c7a8 (update the dependecies)

require (
	github.com/DataDog/datadog-agent/test/fakeintake v0.72.2
	github.com/DataDog/datadog-agent/test/new-e2e v0.72.2
	github.com/DataDog/test-infra-definitions v0.0.6-0.20251114113621-929e070e7069
	github.com/google/go-cmp v0.7.0
<<<<<<< HEAD
	github.com/gruntwork-io/terratest v0.46.16
	github.com/pulumi/pulumi/sdk/v3 v3.190.0
=======
go 1.24.0

toolchain go1.24.4

require (
	github.com/DataDog/datadog-agent/test/fakeintake v0.67.0
	github.com/DataDog/datadog-agent/test/new-e2e v0.70.0-devel.0.20250723220959-40fc15132396
	github.com/DataDog/test-infra-definitions v0.0.4-0.20250725180812-83c23398aae9
	github.com/google/go-cmp v0.7.0
	github.com/gruntwork-io/terratest v0.47.2
	github.com/pulumi/pulumi/sdk/v3 v3.181.0
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
	github.com/stretchr/testify v1.10.0
=======
	github.com/gruntwork-io/terratest v0.47.2
	github.com/pulumi/pulumi/sdk/v3 v3.190.0
	github.com/stretchr/testify v1.11.1
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.32.3
	k8s.io/apiextensions-apiserver v0.31.1
	k8s.io/apimachinery v0.32.3
)

require (
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/DataDog/agent-payload/v5 v5.0.145 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.65.1 // indirect
	github.com/DataDog/datadog-agent/comp/netflow/payload v0.65.1 // indirect
	github.com/DataDog/datadog-agent/pkg/metrics v0.65.1 // indirect
	github.com/DataDog/datadog-agent/pkg/network/payload v0.65.1 // indirect
	github.com/DataDog/datadog-agent/pkg/networkpath/payload v0.65.1 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.65.1 // indirect
	github.com/DataDog/datadog-agent/pkg/tagger/types v0.65.1 // indirect
=======
	github.com/DataDog/agent-payload/v5 v5.0.150 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.67.0 // indirect
	github.com/DataDog/datadog-agent/comp/netflow/payload v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/metrics v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/network/payload v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/networkpath/payload v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/tagger/types v0.67.0 // indirect
>>>>>>> 7844aff2 (updated the dependecies)
	github.com/DataDog/datadog-agent/pkg/util/option v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.67.0 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.35.0 // indirect
=======
	github.com/DataDog/agent-payload/v5 v5.0.158 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.68.0 // indirect
	github.com/DataDog/datadog-agent/comp/netflow/payload v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/metrics v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/network/payload v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/networkpath/payload v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/tagger/types v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/util/option v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.68.0 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.68.0 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.41.0 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
=======
	github.com/DataDog/agent-payload/v5 v5.0.166 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.72.2 // indirect
	github.com/DataDog/datadog-agent/comp/netflow/payload v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/metrics v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/network/payload v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/networkpath/payload v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/tagger/types v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/util/option v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/util/pointer v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.72.2 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.72.2 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.46.0 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	github.com/DataDog/mmh3 v0.0.0-20210722141835-012dc69a9e49 // indirect
	github.com/DataDog/zstd v1.5.6 // indirect
	github.com/DataDog/zstd_0 v0.0.0-20210310093942-586c1286621f // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/alessio/shellescape v1.4.2 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/aws/aws-sdk-go v1.55.6 // indirect
=======
	github.com/aws/aws-sdk-go v1.55.7 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
	github.com/aws/aws-sdk-go-v2 v1.36.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.70 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.36 // indirect
=======
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.9 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.7 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecr v1.45.1 // indirect
<<<<<<< HEAD
	github.com/aws/aws-sdk-go-v2/service/ecs v1.58.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.83.0 // indirect
<<<<<<< HEAD
	github.com/aws/aws-sdk-go-v2/service/ssm v1.56.12 // indirect
=======
	github.com/aws/aws-sdk-go-v2/service/ssm v1.59.3 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.0 // indirect
=======
	github.com/aws/aws-sdk-go-v2/service/ecs v1.64.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.64.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.4 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	github.com/aws/session-manager-plugin v0.0.0-20241119210807-82dc72922492 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/charmbracelet/bubbles v0.20.0 // indirect
	github.com/charmbracelet/bubbletea v1.2.4 // indirect
	github.com/charmbracelet/colorprofile v0.3.0 // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13 // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/cheggaaa/pb v1.0.29 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/docker/cli v27.5.0+incompatible // indirect
	github.com/docker/docker v28.4.0+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/go-git/go-billy/v5 v5.6.1 // indirect
	github.com/go-git/go-git/v5 v5.13.1 // indirect
<<<<<<< HEAD
	github.com/go-logr/logr v1.4.2 // indirect
=======
=======
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.13.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
=======
	github.com/go-logr/logr v1.4.3 // indirect
>>>>>>> 7844aff2 (updated the dependecies)
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-sql-driver/mysql v1.8.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gonvenience/bunt v1.3.5 // indirect
	github.com/gonvenience/neat v1.3.12 // indirect
	github.com/gonvenience/term v1.0.2 // indirect
	github.com/gonvenience/text v1.0.7 // indirect
	github.com/gonvenience/wrap v1.1.2 // indirect
	github.com/gonvenience/ytbx v1.4.4 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/gruntwork-io/go-commons v0.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl/v2 v2.23.0 // indirect
	github.com/homeport/dyff v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/iwdgo/sigintwindows v0.2.2 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-ciede2000 v0.0.0-20170301095244-782e8c62fec3 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-zglob v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opentracing/basictracer-go v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pgavlin/fx v0.1.6 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.9 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/otp v1.2.0 // indirect
	github.com/pulumi/appdash v0.0.0-20231130102222-75f619a67231 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/pulumi/esc v0.17.0 // indirect
=======
	github.com/pulumi/esc v0.14.3 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
=======
	github.com/pulumi/esc v0.14.3 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
=======
	github.com/pulumi/esc v0.17.0 // indirect
>>>>>>> 0b53c7a8 (update the dependecies)
	github.com/pulumi/pulumi-aws/sdk/v6 v6.66.2 // indirect
	github.com/pulumi/pulumi-awsx/sdk/v2 v2.19.0 // indirect
	github.com/pulumi/pulumi-azure-native-sdk/v2 v2.81.0 // indirect
	github.com/pulumi/pulumi-command/sdk v1.0.1 // indirect
	github.com/pulumi/pulumi-docker/sdk/v4 v4.9.0 // indirect
	github.com/pulumi/pulumi-eks/sdk/v3 v3.7.0 // indirect
	github.com/pulumi/pulumi-gcp/sdk/v7 v7.38.0 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/pulumi/pulumi-kubernetes/sdk/v4 v4.23.0 // indirect
	github.com/pulumi/pulumi-random/sdk/v4 v4.18.4 // indirect
=======
	github.com/pulumi/pulumi-kubernetes/sdk/v4 v4.19.0 // indirect
	github.com/pulumi/pulumi-libvirt/sdk v0.5.4 // indirect
	github.com/pulumi/pulumi-random/sdk/v4 v4.16.8 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
=======
	github.com/pulumi/pulumi-kubernetes/sdk/v4 v4.19.0 // indirect
	github.com/pulumi/pulumi-random/sdk/v4 v4.16.8 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
=======
	github.com/pulumi/pulumi-kubernetes/sdk/v4 v4.23.0 // indirect
	github.com/pulumi/pulumi-random/sdk/v4 v4.18.4 // indirect
>>>>>>> 0b53c7a8 (update the dependecies)
	github.com/pulumi/pulumi-tls/sdk/v4 v4.11.1 // indirect
	github.com/pulumiverse/pulumi-time/sdk v0.1.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/tinylib/msgp v1.4.0 // indirect
	github.com/twinj/uuid v0.0.0-20151029044442-89173bcdda19 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/urfave/cli/v2 v2.27.6 // indirect
	github.com/virtuald/go-ordered-json v0.0.0-20170621173500-b18e6e673d74 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zclconf/go-cty v1.15.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
<<<<<<< HEAD
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250224174004-546df14abb99 // indirect
	google.golang.org/grpc v1.71.1 // indirect
=======
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
>>>>>>> 2461b46e (Add generic QA E2E tests (#1970))
=======
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250425173222-7b384671a197 // indirect
	google.golang.org/grpc v1.72.0 // indirect
>>>>>>> 7844aff2 (updated the dependecies)
	google.golang.org/protobuf v1.36.6 // indirect
=======
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250922171735-9219d122eba9 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
>>>>>>> 9f2ef8e6 (updated go.mod and sum file)
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/zorkian/go-datadog-api.v2 v2.30.0 // indirect
	k8s.io/client-go v0.32.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	lukechampine.com/frand v1.5.1 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.5.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
