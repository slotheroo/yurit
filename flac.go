// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"errors"
	"fmt"
	"io"
	"time"
)

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

//FLACMetadata is a collection of metadata and other useful data from a native
//FLAC container with a Vorbis comment.
type FLACMetadata struct {
	fileType      FileType
	streamInfo    StreamInfo
	pictures      []Picture
	vorbisComment *VorbisComment
}

// ReadFLACTags reads FLAC metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
func ReadFLACTags(r io.ReadSeeker) (*FLACMetadata, error) {
	flac, err := readString(r, 4)
	if err != nil {
		return nil, err
	}
	//Verify that this is a FLAC file
	if flac != "fLaC" {
		return nil, errors.New("expected 'fLaC'")
	}

	m := &FLACMetadata{
		fileType: FLAC,
	}
	err = m.readFLACMetadataBlocks(r)
	if err != nil {
		return nil, err
	}
	return m, nil
}

//readFLACMetadataBlocks iterates through a FLAC file's metadata blocks and reads
//the ones that we are interested in.
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

//readStreamInfoBlock reads the STREAMINFO block from a FLAC file
//https://xiph.org/flac/format.html#metadata_block_streaminfo
func (m *FLACMetadata) readStreamInfoBlock(r io.Reader) error {
	block, err := readBytes(r, 34)
	if err != nil {
		return err
	}
	m.streamInfo.MinBlockSize = getInt(block[0:2])
	m.streamInfo.MaxBlockSize = getInt(block[2:4])
	m.streamInfo.MinFrameSize = getInt(block[4:7])
	m.streamInfo.MaxFrameSize = getInt(block[7:10])
	m.streamInfo.SampleRate = getInt(block[10:13]) >> 4
	m.streamInfo.Channels = ((getInt(block[12:13]) >> 1) & 0x07) + 1
	m.streamInfo.SampleBitrate = ((getInt(block[12:13]) & 0x01) << 4) + (getInt(block[13:14]) >> 4) + 1
	m.streamInfo.TotalSamples = getInt(append([]byte{block[13] & 0x0F}, block[14:18]...))
	m.streamInfo.MD5Signature = block[18:]
	return nil
}

//readVorbisComment reads a Vorbis comment from the corresponding metadata block
//in a FLAC file
//https://xiph.org/flac/format.html#metadata_block_vorbis_comment
func (m *FLACMetadata) readVorbisComment(r io.Reader) error {
	var err error
	m.vorbisComment, err = ReadVorbisComment(r)
	if err != nil {
		return err
	}
	return nil
}

//readPictureBlock reads a FLAC picture metadata block into FLACMetadata
//https://xiph.org/flac/format.html#metadata_block_picture
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

	// We skip width <32>, height <32>, colorDepth <32>, colorsUsed <32>
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
	m.pictures = append(m.pictures, picture)
	return nil
}

func (m FLACMetadata) Album() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Album()
	}
	return ""
}

func (m FLACMetadata) AlbumArtist() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.AlbumArtist()
	}
	return ""
}

func (m FLACMetadata) Artist() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Artist()
	}
	return ""
}

func (m FLACMetadata) Comment() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Comment()
	}
	return ""
}

func (m FLACMetadata) Composer() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Composer()
	}
	return ""
}

func (m FLACMetadata) Disc() (int, int) {
	if m.vorbisComment != nil {
		return m.vorbisComment.Disc()
	}
	return 0, 0
}

func (m FLACMetadata) Duration() time.Duration {
	if m.streamInfo.SampleRate == 0 {
		return time.Duration(0)
	}
	//Calculate track length by dividing total samples by sample rate
	seconds := float64(m.streamInfo.TotalSamples) / float64(m.streamInfo.SampleRate)
	//convert to time.Duration
	return time.Duration(seconds * float64(time.Second))
}

func (m FLACMetadata) FileType() FileType {
	return m.fileType
}

func (m FLACMetadata) Format() Format {
	if m.vorbisComment != nil {
		return m.vorbisComment.Format()
	}
	return UnknownFormat
}

func (m FLACMetadata) Genre() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Genre()
	}
	return ""
}

func (m FLACMetadata) Lyrics() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Lyrics()
	}
	return ""
}

//Picture attempts to return front cover art, else it returns the first picture
//found in the FLAC metadata
func (m FLACMetadata) Picture() *Picture {
	if len(m.pictures) == 0 {
		return nil
	}
	for _, pic := range m.pictures {
		if pic.Type == pictureTypes[0x03] {
			return &pic
		}
	}
	return &m.pictures[0]
}

//Pictures returns ALL pictures found in a FLAC file's metadata.
//https://xiph.org/flac/format.html#metadata_block_picture
func (m FLACMetadata) Pictures() []Picture {
	return m.pictures
}

//StreamInfo returns the data extracted from a FLAC file's stream info metadata
//block. See the StreamInfo struct type for more information.
func (m FLACMetadata) StreamInfo() StreamInfo {
	return m.streamInfo
}

func (m FLACMetadata) Title() string {
	if m.vorbisComment != nil {
		return m.vorbisComment.Title()
	}
	return ""
}

func (m FLACMetadata) Track() (int, int) {
	if m.vorbisComment != nil {
		return m.vorbisComment.Track()
	}
	return 0, 0
}

//VorbisComment returns the information found in a FLAC file's Vorbis comment
//metadata block. The data in this block is sometimes also referred to as FLAC
//tags.
func (m FLACMetadata) VorbisComment() *VorbisComment {
	return m.vorbisComment
}

func (m FLACMetadata) Year() int {
	if m.vorbisComment != nil {
		return m.vorbisComment.Year()
	}
	return 0
}

//StreamInfo holds general information about the FLAC audio stream and is
//extracted from a FLAC file's STREAMINFO metadata block
//https://xiph.org/flac/format.html#metadata_block_streaminfo
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
