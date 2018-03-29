package cc111x

import (
	"github.com/ecc1/radio"
)

var (
	// Ensure that *Radio implements the radio.Interface interface.
	_ radio.Interface = (*Radio)(nil)
)
