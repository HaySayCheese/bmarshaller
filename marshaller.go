package marshalling

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"
)

type Encoder struct {
	tmpBuffer []byte
	payload   *bytes.Buffer

	isReleased bool
}

func NewEncoder() (encoder *Encoder) {
	encoder = &Encoder{
		tmpBuffer: tmp8BytesBuffersPool.Get().([]byte),
		payload:   payloadBuffersPool.Get().(*bytes.Buffer),

		isReleased: true,
	}
	return
}

func (e *Encoder) CollectDataAndReleaseBuffers() (data []byte) {
	data = make([]byte, e.payload.Len())
	copy(data, e.payload.Bytes())

	e.isReleased = true
	e.payload.Reset()
	payloadBuffersPool.Put(e.payload)
	tmp8BytesBuffersPool.Put(e.tmpBuffer)
	return
}

func (e *Encoder) PutString(value string) (err error) {
	return e.PutVariadicDataWith2BytesHeader([]byte(value))
}

func (e *Encoder) PutUint64(value uint64) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	binary.BigEndian.PutUint64(e.tmpBuffer[0:8], value)
	e.payload.Write(e.tmpBuffer[0:8])
	return
}

func (e *Encoder) PutUint32(value uint32) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	binary.BigEndian.PutUint32(e.tmpBuffer[0:4], value)
	e.payload.Write(e.tmpBuffer[0:4])
	return
}

func (e *Encoder) PutUint16(value uint16) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	binary.BigEndian.PutUint16(e.tmpBuffer[0:2], value)
	e.payload.Write(e.tmpBuffer[0:2])
	return
}

func (e *Encoder) PutUint8(value uint8) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	e.tmpBuffer[0] = value
	e.payload.Write(e.tmpBuffer[0:1])
	return
}

func (e *Encoder) PutBigIntWithByteHeader(value *big.Int) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	binData := value.Bytes()
	return e.PutVariadicDataWithByteHeader(binData)
}

func (e *Encoder) PutTimeWithByte8Header(value time.Time) (err error) {
	if e.isReleased == false {
		err = ErrEncoderIsReleased
		return
	}

	binData, err := value.MarshalBinary()
	if err != nil {
		return
	}
	return e.PutVariadicDataWithByteHeader(binData)
}

func (e *Encoder) PutVariadicDataWith2BytesHeader(data []byte) (err error) {
	if len(data) > math.MaxUint16 {
		err = fmt.Errorf("value is too long to be serialized with uint8 header: %w", ErrPayloadIsToLarge)
		return
	}

	binary.BigEndian.PutUint16(e.tmpBuffer[0:2], uint16(len(data)))
	e.payload.Write(e.tmpBuffer[0:2])
	e.payload.Write(data)
	return
}

func (e *Encoder) MarshallVariadicDataWith2BytesHeader(s encoding.BinaryMarshaler) (err error) {
	binaryData, err := s.MarshalBinary()
	if err != nil {
		return
	}

	return e.PutVariadicDataWith2BytesHeader(binaryData)
}

func (e *Encoder) PutVariadicDataWithByteHeader(data []byte) (err error) {
	if len(data) > math.MaxUint8 {
		err = fmt.Errorf("value is too long to be serialized with uint8 header: %w", ErrPayloadIsToLarge)
		return
	}

	e.payload.WriteByte(uint8(len(data)))
	e.payload.Write(data)
	return
}

func (e *Encoder) MarshallVariadicDataWithByteHeader(s encoding.BinaryMarshaler) (err error) {
	binaryData, err := s.MarshalBinary()
	if err != nil {
		return
	}

	return e.PutVariadicDataWithByteHeader(binaryData)
}

func (e *Encoder) PutFixedSizeDataSegment(data []byte) (err error) {
	if len(data) == 0 {
		err = ErrPayloadIsToSmall
	}

	e.payload.Write(data)
	return
}

func (e *Encoder) MarshallFixedSizeDataSegment(s encoding.BinaryMarshaler) (err error) {
	binaryData, err := s.MarshalBinary()
	if err != nil {
		return
	}

	return e.PutFixedSizeDataSegment(binaryData)
}
