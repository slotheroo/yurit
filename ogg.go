// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	idType      int = 1
	commentType int = 3
)

type OGGMetadata struct {
	VorbisIDHeader VorbisIDHeader
	VorbisComment  VorbisComment
	TotalGranules  int64
}

type VorbisIDHeader struct {
	Version        int
	Channels       int
	SampleRate     int //hertz
	BitrateMax     int
	BitrateNominal int
	BitrateMin     int
	BlockSize0     int
	BlockSize1     int
}

func (m *OGGMetadata) readVorbisIDHeader(r io.ReadSeeker) error {
	version, err := readUint32LittleEndian(r)
	if err != nil {
		return err
	}
	channels, err := readUint(r, 1)
	if err != nil {
		return err
	}
	sampleRate, err := readUint32LittleEndian(r)
	if err != nil {
		return err
	}
	bitrateMax, err := readSignedInt32LittleEndian(r)
	if err != nil {
		return err
	}
	bitrateNom, err := readSignedInt32LittleEndian(r)
	if err != nil {
		return err
	}
	bitrateMin, err := readSignedInt32LittleEndian(r)
	if err != nil {
		return err
	}
	blockSizeByte, err := readBytes(r, 1)
	if err != nil {
		return err
	}
	//Use bits 0-3 to make a uint and use that as an exponent of 2
	blockSize0 := 1 << binary.LittleEndian.Uint16([]byte{blockSizeByte[0] & 0x0F, 0})
	//Use bits 4-7 to make a uint and use that as an exponent of 2
	blockSize1 := 1 << binary.LittleEndian.Uint16([]byte{blockSizeByte[0] >> 4, 0})
	//Skip the framing flag
	_, err = r.Seek(1, io.SeekCurrent)
	if err != nil {
		return err
	}
	m.VorbisIDHeader = VorbisIDHeader{
		Version:        int(version),
		Channels:       int(channels),
		SampleRate:     int(sampleRate),
		BitrateMax:     int(bitrateMax),
		BitrateNominal: int(bitrateNom),
		BitrateMin:     int(bitrateMin),
		BlockSize0:     blockSize0,
		BlockSize1:     blockSize1,
	}
	return nil
}

func (m *OGGMetadata) readVorbisComment(r io.Reader) error {
	comment, err := ReadVorbisComment(r)
	if err != nil {
		return err
	}
	m.VorbisComment = *comment
	return nil
}

func (m *OGGMetadata) getTotalGranules(r io.ReadSeeker) error {
	var err error
	_, err = r.Seek(-14, io.SeekEnd)
	if err != nil {
		return err
	}
	//Start looking backwards for page header capture pattern
	for {
		b, err := readBytes(r, 4)
		if err != nil {
			return err
		}
		if string(b) == "OggS" {
			break
		} else if string(b[:3]) == "ggS" {
			_, err = r.Seek(-5, io.SeekCurrent)
		} else if string(b[:2]) == "gS" {
			_, err = r.Seek(-6, io.SeekCurrent)
		} else if b[0] == 'S' {
			_, err = r.Seek(-7, io.SeekCurrent)
		} else {
			_, err = r.Seek(-8, io.SeekCurrent)
		}
		if err != nil {
			return err
		}
	}
	//Skip version byte
	_, err = r.Seek(1, io.SeekCurrent)
	if err != nil {
		return err
	}
	//Read and check header_type_flag
	headerTypeFlag, err := readBytes(r, 1)
	if err != nil {
		return err
	}
	if headerTypeFlag[0]&0x04 != 0x04 {
		return errors.New("Last page found is not marked as final page")
	}
	//Read final absolute granule position
	absoluteGranulePositionUnsigned, err := readUint64LittleEndian(r)
	if err != nil {
		return err
	}
	m.TotalGranules = int64(absoluteGranulePositionUnsigned)
	return nil
}

// ReadOGGTags reads OGG metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
// See http://www.xiph.org/vorbis/doc/Vorbis_I_spec.html
// and http://www.xiph.org/ogg/doc/framing.html for details.
func ReadOGGTags(r io.ReadSeeker) (*OGGMetadata, error) {
	m := &OGGMetadata{}

	ih, err := readPackets(r)
	if err != nil {
		return nil, err
	}
	ihr := bytes.NewReader(ih)

	// First packet type is identification, type 1
	t, err := readInt(ihr, 1)
	if err != nil {
		return nil, err
	}
	if t != idType {
		return nil, errors.New("expected 'vorbis' identification type 1")
	}

	// Seek and discard 6 bytes from common header
	_, err = ihr.Seek(6, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	err = m.readVorbisIDHeader(ihr)
	if err != nil {
		return nil, err
	}

	// Read comment header packet. May include setup header packet, if it is on the
	// same page. First audio packet is guaranteed to be on the separate page.
	// See https://www.xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-132000A.2
	ch, err := readPackets(r)
	if err != nil {
		return nil, err
	}
	chr := bytes.NewReader(ch)

	// First packet type is comment, type 3
	t, err = readInt(chr, 1)
	if err != nil {
		return nil, err
	}
	if t != commentType {
		return nil, errors.New("expected 'vorbis' comment type 3")
	}

	// Seek and discard 6 bytes from common header
	_, err = chr.Seek(6, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	err = m.readVorbisComment(chr)
	if err != nil {
		return nil, err
	}

	err = m.getTotalGranules(r)
	return m, err
}

// readPackets reads vorbis header packets from contiguous ogg pages in ReadSeeker.
// The pages are considered contiguous, if the first lacing value in second
// page's segment table continues rather than begins a packet. This is indicated
// by setting header_type_flag 0x1 (continued packet).
// See https://www.xiph.org/ogg/doc/framing.html on packets spanning pages.
func readPackets(r io.ReadSeeker) ([]byte, error) {
	buf := &bytes.Buffer{}

	firstPage := true
	for {
		// Read capture pattern
		oggs, err := readString(r, 4)
		if err != nil {
			return nil, err
		}
		if oggs != "OggS" {
			return nil, errors.New("expected 'OggS'")
		}

		// Read page header
		head, err := readBytes(r, 22)
		if err != nil {
			return nil, err
		}
		headerTypeFlag := head[1]

		continuation := headerTypeFlag&0x1 > 0
		if !(firstPage || continuation) {
			// Rewind to the beginning of the page
			_, err = r.Seek(-26, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			break
		}
		firstPage = false

		// Read the number of segments
		nS, err := readUint(r, 1)
		if err != nil {
			return nil, err
		}

		// Read segment table
		segments, err := readBytes(r, nS)
		if err != nil {
			return nil, err
		}

		// Calculate remaining page size
		pageSize := 0
		for i := uint(0); i < nS; i++ {
			pageSize += int(segments[i])
		}

		_, err = io.CopyN(buf, r, int64(pageSize))
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
