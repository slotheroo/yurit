package yurit

import (
	"io"
	"os"
)

//MP3Metadata is a collection of metadata from an mp3 file including tags and
//frame information.
type MP3Metadata struct {
	ID3v2       *ID3v2
	FrameHeader MPEGFrameHeader
	FrameData   MP3FrameData
	ID3v1       MetadataID3v1 //It's a map, not a struct, don't make pointer
}

func ReadFromMP3(file *os.File) (*MP3Metadata, error) {
	var (
		frameHeader MPEGFrameHeader
		frameData   MP3FrameData
	)
	id3v2, err := extractID3v2(file)
	if err != nil {
		return nil, err
	}
	if id3v2 != nil {
		_, err = file.Seek(int64(id3v2.Header.Size), io.SeekStart)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}
	frameHeader, err = readMPEGFrameHeader(file)
	if err != nil {
		return nil, err
	}
	frameData, err = readMP3FrameData(file, frameHeader)
	if err != nil {
		return nil, err
	}
	id3v1, err := ReadID3v1Tags(file)
	if err != nil && err != ErrNotID3v1 {
		return nil, err
	}

	m := MP3Metadata{ID3v2: id3v2, FrameHeader: frameHeader, FrameData: frameData, ID3v1: id3v1}
	return &m, nil
}
