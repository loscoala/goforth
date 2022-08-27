package goforth

import (
  _ "embed"
)

// The core words dictionary
//
//go:embed core.fs
var Core string

// Colored output
var Colored bool

