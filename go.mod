module github.com/ilychi/backtrace

go 1.22.4

toolchain go1.23.2

require (
	github.com/oneclickvirt/defaultset v0.0.0-20240624051018-30a50859e1b5
	golang.org/x/net v0.33.0
	golang.org/x/sys v0.28.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace github.com/nxtrace/NTrace-core => ../NTrace-core
