package sqlite

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// serializeVector converts a float32 slice to a LittleEndian byte slice
// compatible with sqlite-vec BLOB input.
func serializeVector(vec []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, vec)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize vector: %w", err)
	}
	return buf.Bytes(), nil
}
