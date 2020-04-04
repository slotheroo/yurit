// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

//Vorbis common header types (which are always odd numbers)
//https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-620004.2.1
const (
	idType      int = 1
	commentType int = 3
	//setupType int = 5
)

//OggMetadata is a collection of metadata and other useful data from an Ogg
//container that contains Vorbis encoded audio
type OggMetadata struct {
	fileType       FileType
	vorbisIDHeader VorbisIDHeader
	vorbisComment  VorbisComment
	totalGranules  int64
}

// ReadOggTags reads Ogg metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
// See http://www.xiph.org/vorbis/doc/Vorbis_I_spec.html
// and http://www.xiph.org/ogg/doc/framing.html for details.
func ReadOggTags(r io.ReadSeeker) (*OggMetadata, error) {
	m := &OggMetadata{}

	ih, err := readPackets(r)
	if err != nil {
		return nil, err
	}
	ihr := bytes.NewReader(ih)
	m.fileType = OGG

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

//Reads the identification header from a Vorbis audio stream
//See https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-630004.2.2
func (m *OggMetadata) readVorbisIDHeader(r io.ReadSeeker) error {
	//Identification header is 23 bytes long
	b, err := readBytes(r, 23)
	if err != nil {
		return err
	}
	version := getUintLittleEndian(b[0:4])
	channels := b[4]
	sampleRate := getUintLittleEndian(b[5:9])
	bitrateMax := getSignedInt32LittleEndian(b[9:13])
	bitrateNom := getSignedInt32LittleEndian(b[13:17])
	bitrateMin := getSignedInt32LittleEndian(b[17:21])
	//Use bits 0-3 of byte 21 to make a uint and use that as an exponent of 2
	blockSize0 := 1 << binary.LittleEndian.Uint16([]byte{b[21] & 0x0F, 0})
	//Use bits 4-7 of byte 21 to make a uint and use that as an exponent of 2
	blockSize1 := 1 << binary.LittleEndian.Uint16([]byte{b[21] >> 4, 0})
	//Las byte is the framing flag. Ignore.
	m.vorbisIDHeader = VorbisIDHeader{
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

//readVorbisComment reads a Vorbis comment into OggMetadata
func (m *OggMetadata) readVorbisComment(r io.Reader) error {
	comment, err := ReadVorbisComment(r)
	if err != nil {
		return err
	}
	m.vorbisComment = *comment
	return nil
}

//getTotalGranules finds the last page in an Ogg container and extracts the
//absolute granule position, which in this case is the total number of granules
//for the Ogg container
//https://www.xiph.org/ogg/doc/framing.html
func (m *OggMetadata) getTotalGranules(r io.ReadSeeker) error {
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
	m.totalGranules = int64(absoluteGranulePositionUnsigned)
	return nil
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

func (m OggMetadata) Album() string {
	return m.vorbisComment.Album()
}

func (m OggMetadata) AlbumArtist() string {
	return m.vorbisComment.AlbumArtist()
}

func (m OggMetadata) Artist() string {
	return m.vorbisComment.Artist()
}

func (m OggMetadata) Comment() string {
	return m.vorbisComment.Comment()
}

func (m OggMetadata) Composer() string {
	return m.vorbisComment.Composer()
}

func (m OggMetadata) Disc() (int, int) {
	return m.vorbisComment.Disc()
}

//TODO
func (m OggMetadata) Duration() time.Duration {
	if m.vorbisIDHeader.SampleRate == 0 {
		return time.Duration(0)
	}
	//Calculate track length by dividing total samples (granules) by sample rate
	seconds := float64(m.totalGranules) / float64(m.vorbisIDHeader.SampleRate)
	//convert to time.Duration
	return time.Duration(seconds * float64(time.Second))
}

func (m OggMetadata) FileType() FileType {
	return m.fileType
}

func (m OggMetadata) Format() Format {
	return m.vorbisComment.Format()
}

func (m OggMetadata) Genre() string {
	return m.vorbisComment.Genre()
}

func (m OggMetadata) Lyrics() string {
	return m.vorbisComment.Lyrics()
}

//Picture for OggMetadata always returns nil.
//There is no standard location for pictures in an Ogg container unless they are
//muxed into a separate stream, which this library does not handle.
func (m OggMetadata) Picture() *Picture {
	return nil
}

func (m OggMetadata) Title() string {
	return m.vorbisComment.Title()
}

//Returns the total number of granules in this Ogg container
func (m OggMetadata) TotalGranules() int64 {
	return m.totalGranules
}

func (m OggMetadata) Track() (int, int) {
	return m.vorbisComment.Track()
}

//VorbisComment returns the Vorbis comment information associated with this Ogg
//file.
func (m OggMetadata) VorbisComment() VorbisComment {
	return m.vorbisComment
}

//VorbisIDHeader returns the Vorbis identification header information associated
//with this Ogg file. See the VorbisIDHeader struct type for more information.
func (m OggMetadata) VorbisIDHeader() VorbisIDHeader {
	return m.vorbisIDHeader
}

func (m OggMetadata) Year() int {
	return m.vorbisComment.Year()
}

//VorbisIDHeader holds general information about a Vorbis audio stream.
//https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-630004.2.2
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
