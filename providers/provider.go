package providers

import (
	"errors"
	"io"
)

// Common errors providers can use
var (
	ErrBadOptions = errors.New("options are invalid or unrecognized")
)

var Providers []Provider

var NameToProvider = map[string]Provider{}

// Provider is the storage provider interface.
type Provider interface {
	// Name is the lowercase ASCII name with no spaces
	Name() string

	Capabilities() Capabilities

	// Upload takes a byte stream and returns an access code for the file or an error.
	// The options are custom and provider-specific.
	Upload(r io.Reader, options map[string]interface{}) (string, error)

	// FileStatus takes a file access code and returns information about the file.
	// An error should be returned if the provider is inaccessible or something
	// similar, not if the file doesn't exist.
	FileInfo(string) (*FileInfo, error)
}

// ProvideRemover is a provider that can remove content.
type ProvideRemover interface {
	Provider

	// Remove takes a file access code and returns an error if it failed,
	// or nil if it succeeded.
	Remove(string) error
}

type FileInfo struct {
	// File is known by the provider or not
	Exists bool
	// File is past the processing stage of the provider
	DoneProcessing bool
	// Custom holds provider-specific information
	Custom map[string]interface{}
}

// Capabilities defines the features a provider has.
type Capabilities struct {
	Removal    bool
	Geofencing bool
}
