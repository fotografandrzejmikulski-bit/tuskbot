//go:build darwin && arm64

package llamacpp

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -lllama -lggml -lggml-cpu -lggml-base -lc++
import "C"
