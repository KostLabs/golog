module benchmarks

go 1.25.1

require (
	github.com/KostLabs/golog v0.0.0
	github.com/apex/log v1.9.0
	github.com/rs/zerolog v1.34.0
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.0
)

// Use local golog module with relative path
replace github.com/KostLabs/golog => ../

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
