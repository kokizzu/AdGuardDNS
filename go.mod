module github.com/AdguardTeam/AdGuardDNS

go 1.24.5

require (
	// NOTE:  Do not change the pseudoversion.
	github.com/AdguardTeam/AdGuardDNS/internal/dnsserver v0.0.0-00010101000000-000000000000
	github.com/AdguardTeam/golibs v0.32.15
	github.com/AdguardTeam/urlfilter v0.20.0
	github.com/ameshkov/dnscrypt/v2 v2.4.0
	github.com/axiomhq/hyperloglog v0.2.5
	github.com/bluele/gcache v0.0.2
	github.com/c2h5oh/datasize v0.0.0-20231215233829-aa82cc1e6500
	github.com/caarlos0/env/v7 v7.1.0
	github.com/getsentry/sentry-go v0.34.0
	github.com/gomodule/redigo v1.9.2
	github.com/google/renameio/v2 v2.0.0
	github.com/miekg/dns v1.1.66
	github.com/oschwald/maxminddb-golang v1.13.1
	github.com/patrickmn/go-cache v2.1.1-0.20191004192108-46f407853014+incompatible
	github.com/prometheus/client_golang v1.22.0
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.64.0
	github.com/quic-go/quic-go v0.52.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.39.0
	golang.org/x/net v0.41.0
	golang.org/x/sys v0.34.0
	golang.org/x/time v0.12.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.121.3 // indirect
	cloud.google.com/go/ai v0.12.1 // indirect
	cloud.google.com/go/auth v0.16.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/ameshkov/dnsstamps v1.0.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/caarlos0/env/v11 v11.3.1 // indirect
	github.com/ccojocar/zxcvbn-go v1.0.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fzipp/gocyclo v0.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/golangci/misspell v0.7.0 // indirect
	github.com/google/generative-ai-go v0.20.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250607225305-033d6d78b36a // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.2 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/gordonklaus/ineffassign v0.1.0 // indirect
	github.com/jstemmer/go-junit-report/v2 v2.1.0 // indirect
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/kisielk/errcheck v1.9.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/ginkgo/v2 v2.23.4 // indirect
	github.com/panjf2000/ants/v2 v2.11.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/securego/gosec/v2 v2.22.5 // indirect
	github.com/uudashr/gocognit v1.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.62.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/mock v0.5.2 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/exp/typeparams v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/telemetry v0.0.0-20250710130107-8d8967aff50b // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	google.golang.org/api v0.241.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.6.1 // indirect
	mvdan.cc/editorconfig v0.3.0 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
	mvdan.cc/sh/v3 v3.12.0 // indirect
	mvdan.cc/unparam v0.0.0-20250301125049-0df0534333a4 // indirect
)

tool (
	github.com/fzipp/gocyclo/cmd/gocyclo
	github.com/golangci/misspell/cmd/misspell
	github.com/gordonklaus/ineffassign
	github.com/jstemmer/go-junit-report/v2
	github.com/kisielk/errcheck
	github.com/securego/gosec/v2/cmd/gosec
	github.com/uudashr/gocognit/cmd/gocognit
	golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
	golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness
	golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	golang.org/x/vuln/cmd/govulncheck
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
	google.golang.org/protobuf
	google.golang.org/protobuf/cmd/protoc-gen-go
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
	mvdan.cc/sh/v3/cmd/shfmt
	mvdan.cc/unparam
)

replace github.com/AdguardTeam/AdGuardDNS/internal/dnsserver => ./internal/dnsserver
