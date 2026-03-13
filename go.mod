module goldbox-rpg

go 1.25.6

toolchain go1.25.8

require (
	github.com/coder/websocket v1.8.14
	github.com/getkin/kin-openapi v0.133.0
	github.com/google/uuid v1.6.0
	// gorilla/websocket is used only for E2E tests (test/e2e/client.go) and benchmarks
	// (pkg/server/benchmark_test.go) as a WebSocket client library. Production code uses
	// github.com/coder/websocket (maintained fork of nhooyr.io/websocket). The archived
	// status of gorilla/websocket is acceptable for test-only usage with fixed v1.5.3.
	github.com/gorilla/websocket v1.5.3
	github.com/hajimehoshi/ebiten/v2 v2.7.0
	github.com/mb-14/gomarkov v0.0.0-20231120193207-9cbdc8df67a8
	github.com/prometheus/client_golang v1.23.2
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.11.1
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/time v0.12.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ebitengine/gomobile v0.0.0-20240329170434-1771503ff0a8 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/purego v0.7.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/jezek/xgb v1.1.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/woodsbury/decimal128 v1.3.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
