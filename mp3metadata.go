package yurit

import (
	"io"
	"os"
)

//MP3Metadata is a collection of metadata from an mp3 file including tags and
//frame information.
type MP3Metadata struct {
	ID3v2Tags   *ID3v2Tags
	FrameHeader MPEGFrameHeader
	FrameData   MP3FrameData
	ID3v1Tags   *ID3v1Tags
}

func (m MP3Metadata) Format() Format {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Format()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Format()
	}
	return UnknownFormat
}

func (m MP3Metadata) FileType() FileType {
	if m.FrameHeader.Layer == MPEGLayer3 {
		return MP3
	} else if m.FrameHeader.Layer == MPEGLayer2 {
		return MP2
	} else if m.FrameHeader.Layer == MPEGLayer1 {
		return MP1
	}
	return UnknownFileType
}

func (m MP3Metadata) Title() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Title()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Title()
	}
	return ""
}

func (m MP3Metadata) Album() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Album()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Album()
	}
	return ""
}

func (m MP3Metadata) Artist() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Artist()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Artist()
	}
	return ""
}

func (m MP3Metadata) AlbumArtist() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.AlbumArtist()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Composer() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Composer()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Genre() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Genre()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Genre()
	}
	return ""
}

func (m MP3Metadata) Year() int {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Year()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Year()
	}
	return 0
}

func (m MP3Metadata) Track() (int, int) {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Track()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Track()
	}
	return 0, 0
}

func (m MP3Metadata) Disc() (int, int) {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Disc()
	}
	//No equivalent value for ID3v1
	return 0, 0
}

func (m MP3Metadata) Lyrics() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Lyrics()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Comment() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Comment()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Comment()
	}
	return ""
}

func (m MP3Metadata) Picture() *Picture {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Picture()
	}
	//No equivalent value for ID3v1
	return nil
}

func ReadFromMP3(file *os.File) (*MP3Metadata, error) {
	var (
		frameHeader MPEGFrameHeader
		frameData   MP3FrameData
	)
	//Extract any ID3v2 tags, if any
	id3v2, err := ReadID3v2Tags(file)
	if err != nil {
		return nil, err
	}
	//Seek to the end of the ID3v2 tags, or the beginning of the file if there are
	//no tags.
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
	//Find and read the first encountered frame header
	frameHeader, err = readMPEGFrameHeader(file)
	if err != nil {
		return nil, err
	}
	//Read frame data immediately after header, seek should be in correct position
	frameData, err = readMP3FrameData(file, frameHeader)
	if err != nil {
		return nil, err
	}
	//Look for an ID3v1 tag at the end of the file
	id3v1, err := ReadID3v1Tags(file)
	if err != nil && err != ErrNotID3v1 {
		return nil, err
	}

	m := MP3Metadata{ID3v2Tags: id3v2, FrameHeader: frameHeader, FrameData: frameData, ID3v1Tags: id3v1}
	return &m, nil
}
