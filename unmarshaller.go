package marshalling

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"time"
)

type Decoder struct {
	reader io.Reader
}

func NewDecoder(data []byte) (decoder *Decoder) {
	return &Decoder{
		reader: bytes.NewReader(data),
	}
}

func NewDecoderFromReader(reader io.Reader) (decoder *Decoder) {
	return &Decoder{
		reader: bufio.NewReader(reader),
	}
}

func (d *Decoder) GetString() (value string, err error) {
	data, err := d.GetDataSegmentWith2BytesHeader()
	if err != nil {
		return
	}

	value = string(data)
	return
}

func (d *Decoder) GetUint64() (value uint64, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:8])
	if err != nil {
		return
	}

	value = binary.BigEndian.Uint64(tmp)
	return
}

func (d *Decoder) GetUint32() (value uint32, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:4])
	if err != nil {
		return
	}

	value = binary.BigEndian.Uint32(tmp)
	return
}

func (d *Decoder) GetUint16() (value uint16, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:2])
	if err != nil {
		return
	}

	value = binary.BigEndian.Uint16(tmp)
	return
}

func (d *Decoder) GetUint8() (value uint8, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:1])
	if err != nil {
		return
	}

	value = tmp[0]
	return
}

func (d *Decoder) GetBigIntWithByteHeader() (value *big.Int, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:1])
	valueSize := int(tmp[0])
	if err != nil {
		return
	}

	data, err := d.readDataSegment(valueSize)
	if err != nil {
		return
	}

	value = &big.Int{}
	value.SetBytes(data)
	return
}

func (d *Decoder) GetTimeWithUint8Header() (value time.Time, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	_, err = d.reader.Read(tmp[0:1])
	valueSize := int(tmp[0])
	if err != nil {
		return
	}

	data, err := d.readDataSegment(valueSize)
	if err != nil {
		return
	}

	err = value.UnmarshalBinary(data)
	return
}

func (d *Decoder) GetDataSegment(size int) (data []byte, err error) {
	return d.readDataSegment(size)
}

func (d *Decoder) UnmarshalDataSegment(size int, destination encoding.BinaryUnmarshaler) (err error) {
	data, err := d.GetDataSegment(size)
	if err != nil {
		return
	}

	err = destination.UnmarshalBinary(data)
	return
}

func (d *Decoder) GetDataSegmentWithByteHeader() (data []byte, err error) {
	dataLength, err := d.GetUint8()
	if err != nil {
		return
	}

	return d.readDataSegment(int(dataLength))
}

func (d *Decoder) UnmarshalDataSegmentWithByteHeader(destination encoding.BinaryUnmarshaler) (err error) {
	data, err := d.GetDataSegmentWithByteHeader()
	if err != nil {
		return
	}

	err = destination.UnmarshalBinary(data)
	return
}

func (d *Decoder) GetDataSegmentWith2BytesHeader() (data []byte, err error) {
	dataLength, err := d.GetUint16()
	if err != nil {
		return
	}

	return d.readDataSegment(int(dataLength))
}

func (d *Decoder) UnmarshalDataSegmentWith2BytesHeader(destination encoding.BinaryUnmarshaler) (err error) {
	data, err := d.GetDataSegmentWith2BytesHeader()
	if err != nil {
		return
	}

	err = destination.UnmarshalBinary(data)
	return
}

func (d *Decoder) readDataSegment(size int) (data []byte, err error) {
	tmp := tmp8BytesBuffersPool.Get().([]byte)
	defer tmp8BytesBuffersPool.Put(tmp)

	buffer := payloadBuffersPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		payloadBuffersPool.Put(buffer)
	}()

	bytesProcessed := 0
	for {
		if bytesProcessed == size {
			break
		}

		delta := size - bytesProcessed
		if delta >= 8 {
			i, err := d.reader.Read(tmp)
			if err != nil || i != 8 {
				err = fmt.Errorf("can't read next 8 bytes of data. details: %w", err)
				return nil, err
			}

			bytesProcessed += i
			buffer.Write(tmp)

		} else {
			i, err := d.reader.Read(tmp[0:delta])
			if err != nil || i != delta {
				err = fmt.Errorf("can't read next %d bytes of data. details: %w", delta, err)
				return nil, err
			}

			bytesProcessed += i
			buffer.Write(tmp[0:delta])
		}
	}

	bytesData := buffer.Bytes()
	data = make([]byte, len(bytesData))
	copy(data, bytesData[:])

	return
}
