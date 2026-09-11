package version

// Name is the CLI binary name.
var Name = "mcae"

// Version is the current release version.
// It can be overridden at build time with:
//
//	go build -ldflags "-X github.com/tidjee-dev/mcae/internal/version.Version=x.y.z"
var Version = "0.1.0"
