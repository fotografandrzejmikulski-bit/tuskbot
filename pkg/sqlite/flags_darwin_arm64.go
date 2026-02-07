//go:build darwin && arm64

package sqlite

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -lsqlite_vec -lm
import "C"
