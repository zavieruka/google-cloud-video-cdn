// Package version holds the application's build version.
package version

// Version is the build version, overridden at link time via:
//
//	-ldflags "-X github.com/zavieruka/video-platform/backend/internal/version.Version=<value>"
//
// It defaults to "dev" for local builds.
var Version = "dev"
