package yurit

import (
	"errors"
	"fmt"
	"io"
	"os"
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
	vorbisCommentBlock blockType = 4
	// Cue Sheet Block             5
	pictureBlock blockType = 6
)

//FLACMetadata is a collection of metadata and other useful data from a native
//FLAC container with a Vorbis comment.
type FLACMetadata struct {
	fileType      FileType
	fileSize      int64
	metadataSize  int64
	streamInfo    flacStreamInfo
	pictures      []Picture
	vorbisComment vorbisComment
}

// ReadFLACTags reads FLAC metadata from a FLAC file, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
func ReadFLACTags(file *os.File) (*FLACMetadata, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	flac, err := readString(file, 4)
	if err != nil {
		return nil, err
	}
	//Verify that this is a FLAC file
	if flac != "fLaC" {
		return nil, errors.New("expected 'fLaC'")
	}

	m := &FLACMetadata{
		fileType:     FLAC,
		fileSize:     stat.Size(),
		metadataSize: 4, //fLaC
	}
	err = m.readFLACMetadataBlocks(file)
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
		m.metadataSize += int64(blockLen)

		if blockType(blockHeader[0]) == streamInfoBlock {
			b, err := readBytes(r, uint(blockLen))
			if err != nil {
				return err
			}
			err = m.loadStreamInfo(b)
		} else if blockType(blockHeader[0]) == vorbisCommentBlock {
			//m.vorbisComment, err = readVorbisComment(r)
			b, err := readBytes(r, uint(blockLen))
			if err != nil {
				return err
			}
			err = m.loadVorbisComment(b)
		} else if blockType(blockHeader[0]) == pictureBlock {
			b, err := readBytes(r, uint(blockLen))
			if err != nil {
				return err
			}
			err = m.processLoadPictureBlock(b)
		} else {
			_, err = r.Seek(int64(blockLen), io.SeekCurrent)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//loadStreamInfo processes and loads a stream information from the corresponding
//metadata block in a FLAC file
//https://xiph.org/flac/format.html#metadata_block_streaminfo
func (m *FLACMetadata) loadStreamInfo(b []byte) error {
	si, err := processStreamInfoBlock(b)
	m.streamInfo = si
	return err
}

//loadVorbisComment processes and loads a Vorbis comment from the corresponding
//metadata block in a FLAC file
//https://xiph.org/flac/format.html#metadata_block_vorbis_comment
func (m *FLACMetadata) loadVorbisComment(b []byte) error {
	vc, err := processVorbisComment(b)
	m.vorbisComment = vc
	return err
}

//processLoadPictureBlock processes a FLAC picture metadata block and loads the
//picture into the FLACMetadata.
//https://xiph.org/flac/format.html#metadata_block_picture
func (m *FLACMetadata) processLoadPictureBlock(b []byte) error {
	if len(b) < 32 {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 32, len(b))
	}
	pictureTypeBytes := b[0:4]
	offset := 4
	pictureType, ok := pictureTypes[pictureTypeBytes[3]]
	if !ok {
		return fmt.Errorf("invalid picture type: %v", pictureTypeBytes)
	}

	mimeLen := int(getUint32AsInt64(b[offset : offset+4])) //mime length
	offset += 4
	if len(b) < 32+mimeLen {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 32+mimeLen, len(b))
	}
	mime := string(b[offset : offset+mimeLen])
	offset += mimeLen
	ext := ""
	switch mime {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	}

	descLen := int(getUint32AsInt64(b[offset : offset+4])) //description length
	offset += 4
	if len(b) < 32+mimeLen+descLen {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 32+mimeLen+descLen, len(b))
	}
	desc := string(b[offset : offset+descLen])
	offset += descLen

	// We skip 16 bytes: width <32>, height <32>, colorDepth <32>, colorsUsed <32>
	offset += 16

	dataLen := int(getUint32AsInt64(b[offset : offset+4])) //data length
	offset += 4
	if len(b) < 32+mimeLen+descLen+dataLen {
		return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 32+mimeLen+descLen+dataLen, len(b))
	}
	data := b[offset : offset+dataLen]

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

// AverageBitrate returns the roughly calculated average bitrate of the file in
// bits per second.
//
// While metadata is discounted for this calculation, frame headers are not,
// so the returned value is likely to be slightly higher than the actual. This
// difference is expected to be minor in most cases, though, and since average
// bitrate for a FLAC file is fairly meaningless, the returned value is
// considered sufficiently accurate.
func (m FLACMetadata) AverageBitrate() int {
	durationInSeconds := m.Duration().Seconds()
	if durationInSeconds == 0 {
		return 0
	}
	//calculate audioDataSize in bits, convert to float64 to use with durationInSeconds
	audioDataSize := float64((m.fileSize - m.metadataSize) * 8)
	return int(audioDataSize / durationInSeconds)
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
	return m.streamInfo.Duration()
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

//SampleRate returns the SampleRate from a FLAC file's stream info block
func (m FLACMetadata) SampleRate() int {
	return m.streamInfo.SampleRate()
}

//StreamInfo returns the data extracted from a FLAC file's stream info metadata
//block. See the StreamInfo struct type for more information.
func (m FLACMetadata) StreamInfo() map[string]interface{} {
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
func (m FLACMetadata) VorbisComment() map[string]string {
	return m.vorbisComment
}

func (m FLACMetadata) Year() int {
	if m.vorbisComment != nil {
		return m.vorbisComment.Year()
	}
	return 0
}
