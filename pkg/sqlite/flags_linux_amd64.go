//go:build linux && amd64

package sqlite

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/linux_amd64 -lsqlite_vec -lm
import "C"
