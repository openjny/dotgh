// Package version provides version information for the application.
// These variables are set at build time using -ldflags.
package version

var (
	// Version is the application version (set at build time).
	Version = "dev"
	// Commit is the git commit SHA (set at build time).
	Commit = "none"
	// Date is the build date in RFC3339 format (set at build time).
	Date = "unknown"
)
