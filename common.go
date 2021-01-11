package marshalling

import (
	"bytes"
	"errors"
	"sync"
)

var (
	ErrEncoderIsReleased = errors.New("encoder has been released and could not accept more data")
	ErrPayloadIsToLarge  = errors.New("payload is too large")
	ErrPayloadIsToSmall  = errors.New("payload is too small")
)

var (
	tmp8BytesBuffersPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 8)
		},
	}
	payloadBuffersPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)
