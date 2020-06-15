// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"bytes"
	"errors"
	"io"
	"time"
)

//Vorbis common header types (which are always odd numbers)
//https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-620004.2.1
const (
	vorbisPacketIDType      byte = 1
	vorbisPacketCommentType byte = 3
	//vorbisPacketSetupType byte = 5
)

//OggMetadata is a collection of metadata and other useful data from an Ogg
//container that contains Vorbis encoded audio
type OggMetadata struct {
	fileType       FileType
	vorbisIDHeader vorbisIDHeader
	vorbisComment  vorbisComment
	totalGranules  int64
}

// ReadOggTags reads Ogg metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
// See http://www.xiph.org/vorbis/doc/Vorbis_I_spec.html
// and http://www.xiph.org/ogg/doc/framing.html for details.
func ReadOggTags(r io.ReadSeeker) (*OggMetadata, error) {
	m := &OggMetadata{}

	idHeaderPacket, err := readPackets(r)
	if err != nil {
		return nil, err
	}
	m.fileType = OGG
	if idHeaderPacket[0] != vorbisPacketIDType {
		return nil, errors.New("expected 'vorbis' identification type 1")
	}
	if string(idHeaderPacket[1:7]) != "vorbis" {
		return nil, errors.New("expected 'vorbis' identifier in identification common header")
	}
	err = m.loadVorbisIDHeader(idHeaderPacket[7:])
	if err != nil {
		return nil, err
	}

	// Read comment header packet. May include setup header packet, if it is on the
	// same page. First audio packet is guaranteed to be on the separate page.
	// See https://www.xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-132000A.2
	commentHeaderPacket, err := readPackets(r)
	if err != nil {
		return nil, err
	}
	//////
	if commentHeaderPacket[0] != vorbisPacketCommentType {
		return nil, errors.New("expected 'vorbis' comment type 3")
	}
	if string(commentHeaderPacket[1:7]) != "vorbis" {
		return nil, errors.New("expected 'vorbis' identifier in comment common header")
	}
	err = m.loadVorbisComment(commentHeaderPacket[7:])
	if err != nil {
		return nil, err
	}

	err = m.getTotalGranules(r)
	return m, err
}

func (m *OggMetadata) loadVorbisIDHeader(b []byte) error {
	vc, err := processVorbisIDHeader(b)
	m.vorbisIDHeader = vc
	return err
}

func (m *OggMetadata) loadVorbisComment(b []byte) error {
	vc, err := processVorbisComment(b)
	m.vorbisComment = vc
	return err
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
	absoluteGranulePosition, err := readInt64Little(r)
	if err != nil {
		return err
	}
	m.totalGranules = absoluteGranulePosition
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

func (m OggMetadata) AverageBitrate() int {
	return m.vorbisIDHeader.AverageBitrate()
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

func (m OggMetadata) Duration() time.Duration {
	return m.vorbisIDHeader.Duration(m.totalGranules)
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
func (m OggMetadata) VorbisComment() map[string]string {
	return m.vorbisComment
}

//VorbisIDHeader returns the Vorbis identification header information associated
//with this Ogg file. See the vorbisIDHeader struct type for more information.
func (m OggMetadata) VorbisIDHeader() map[string]interface{} {
	return m.vorbisIDHeader
}

func (m OggMetadata) Year() int {
	return m.vorbisComment.Year()
}
