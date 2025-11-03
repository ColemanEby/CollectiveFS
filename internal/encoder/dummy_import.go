package encoder

// This file ensures github.com/klauspost/reedsolomon stays in go.mod
// even though it's only used by lib/encoder.go which has //go:build ignore
import (
	_ "github.com/klauspost/reedsolomon"
)
