// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"errors"
	"fmt"
	"io"
)

//MP3Metadata is a collection of metadata from an mp3 file including tags and
//frame information.
type FLACMetadata struct {
	StreamInfo    StreamInfo
	Pictures      []Picture
	VorbisComment *VorbisComment
}

func (m FLACMetadata) FileType() FileType {
	return FLAC
}

func (m FLACMetadata) Format() Format {
	if m.VorbisComment != nil {
		return m.VorbisComment.Format()
	}
	return UnknownFormat
}

func (m FLACMetadata) Title() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Title()
	}
	return ""
}

func (m FLACMetadata) Artist() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Artist()
	}
	return ""
}

func (m FLACMetadata) Album() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Album()
	}
	return ""
}

func (m FLACMetadata) AlbumArtist() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.AlbumArtist()
	}
	return ""
}

func (m FLACMetadata) Composer() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Composer()
	}
	return ""
}

func (m FLACMetadata) Genre() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Genre()
	}
	return ""
}

func (m FLACMetadata) Year() int {
	if m.VorbisComment != nil {
		return m.VorbisComment.Year()
	}
	return 0
}

func (m FLACMetadata) Track() (int, int) {
	if m.VorbisComment != nil {
		return m.VorbisComment.Track()
	}
	return 0, 0
}

func (m FLACMetadata) Disc() (int, int) {
	if m.VorbisComment != nil {
		return m.VorbisComment.Disc()
	}
	return 0, 0
}

func (m FLACMetadata) Lyrics() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Lyrics()
	}
	return ""
}

func (m FLACMetadata) Comment() string {
	if m.VorbisComment != nil {
		return m.VorbisComment.Comment()
	}
	return ""
}

type StreamInfo struct {
	MinBlockSize  int //Min block size in samples
	MaxBlockSize  int //Max block size in samples
	MinFrameSize  int //Min frame size in bytes
	MaxFrameSize  int //Max frame size in bytes
	SampleRate    int //Sample rate in hertz
	Channels      int //Number of channels in the stream
	SampleBitrate int //Bits per sample
	TotalSamples  int //Total number of samples in the stream
	MD5Signature  []byte
}

// blockType is a type which represents an enumeration of valid FLAC blocks
type blockType byte

// FLAC block types.
const (
	streamInfoBlock blockType = 0
	// Padding Block               1
	// Application Block           2
	// Seektable Block             3
	// Cue Sheet Block             5
	vorbisCommentBlock blockType = 4
	pictureBlock       blockType = 6
)

// ReadFLACTags reads FLAC metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
func ReadFLACTags(r io.ReadSeeker) (*FLACMetadata, error) {
	flac, err := readString(r, 4)
	if err != nil {
		return nil, err
	}
	if flac != "fLaC" {
		return nil, errors.New("expected 'fLaC'")
	}

	m := &FLACMetadata{}
	err = m.readFLACMetadataBlocks(r)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *FLACMetadata) readFLACMetadataBlocks(r io.ReadSeeker) error {
	var last = false
	for !last {
		blockHeader, err := readBytes(r, 1)
		if err != nil {
			return err
		}

		if getBit(blockHeader[0], 7) {
			blockHeader[0] ^= (1 << 7)
			last = true
		}
		blockLen, err := readInt(r, 3)
		if err != nil {
			return err
		}
		fmt.Println(blockHeader[0])

		if blockType(blockHeader[0]) == streamInfoBlock {
			err = m.readStreamInfoBlock(r)
		} else if blockType(blockHeader[0]) == vorbisCommentBlock {
			err = m.readVorbisComment(r)
		} else if blockType(blockHeader[0]) == pictureBlock {
			err = m.readPictureBlock(r)
		} else {
			_, err = r.Seek(int64(blockLen), io.SeekCurrent)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *FLACMetadata) readStreamInfoBlock(r io.Reader) error {
	block, err := readBytes(r, 34)
	if err != nil {
		return err
	}
	m.StreamInfo.MinBlockSize = getInt(block[0:2])
	m.StreamInfo.MaxBlockSize = getInt(block[2:4])
	m.StreamInfo.MinFrameSize = getInt(block[4:7])
	m.StreamInfo.MaxFrameSize = getInt(block[7:10])
	m.StreamInfo.SampleRate = getInt(block[10:13]) >> 4
	m.StreamInfo.Channels = ((getInt(block[12:13]) >> 1) & 0x07) + 1
	m.StreamInfo.SampleBitrate = ((getInt(block[12:13]) & 0x01) << 4) + (getInt(block[13:14]) >> 4) + 1
	m.StreamInfo.TotalSamples = getInt(append([]byte{block[13] & 0x0F}, block[14:18]...))
	m.StreamInfo.MD5Signature = block[18:]
	return nil
}

func (m *FLACMetadata) readVorbisComment(r io.Reader) error {
	var err error
	m.VorbisComment, err = ReadVorbisComment(r)
	if err != nil {
		return err
	}
	return nil
}

func (m *FLACMetadata) readPictureBlock(r io.Reader) error {
	b, err := readInt(r, 4)
	if err != nil {
		return err
	}
	pictureType, ok := pictureTypes[byte(b)]
	if !ok {
		return fmt.Errorf("invalid picture type: %v", b)
	}
	mimeLen, err := readUint(r, 4)
	if err != nil {
		return err
	}
	mime, err := readString(r, mimeLen)
	if err != nil {
		return err
	}

	ext := ""
	switch mime {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	}

	descLen, err := readUint(r, 4)
	if err != nil {
		return err
	}
	desc, err := readString(r, descLen)
	if err != nil {
		return err
	}

	// We skip width <32>, height <32>, colorDepth <32>, coloresUsed <32>
	_, err = readInt(r, 4) // width
	if err != nil {
		return err
	}
	_, err = readInt(r, 4) // height
	if err != nil {
		return err
	}
	_, err = readInt(r, 4) // color depth
	if err != nil {
		return err
	}
	_, err = readInt(r, 4) // colors used
	if err != nil {
		return err
	}

	dataLen, err := readInt(r, 4)
	if err != nil {
		return err
	}
	data := make([]byte, dataLen)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return err
	}

	picture := Picture{
		Ext:         ext,
		MIMEType:    mime,
		Type:        pictureType,
		Description: desc,
		Data:        data,
	}
	m.Pictures = append(m.Pictures, picture)
	return nil
}
