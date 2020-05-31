package yurit

import (
	"io"
	"os"
	"time"
)

//MP3Metadata is a collection of metadata from an mp3 file including tags and
//frame information.
type MP3Metadata struct {
	ID3v2Tags   *ID3v2Tags
	fileSize    int64
	FrameHeader MP3FrameHeader
	FrameData   MP3FrameData
	ID3v1Tags   *ID3v1Tags
}

func ReadFromMP3(file *os.File) (*MP3Metadata, error) {
	var (
		frameHeader MP3FrameHeader
		frameData   MP3FrameData
	)
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
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
	frameHeader, err = readMP3FrameHeader(file)
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
	if err != nil {
		return nil, err
	}

	m := MP3Metadata{ID3v2Tags: id3v2, fileSize: stat.Size(), FrameHeader: frameHeader, FrameData: frameData, ID3v1Tags: id3v1}
	return &m, nil
}

// return the approximate size of audio data in bytes
func (m MP3Metadata) approximateAudioSize() int64 {
	var v1TagSize int64 = 0
	var v2TagSize int64 = 0
	if m.ID3v1Tags != nil {
		v1TagSize = 128
	}
	if m.ID3v2Tags != nil {
		v2TagSize = 10 + int64(m.ID3v2Tags.Header.Size)
	}
	return m.fileSize - v1TagSize - v2TagSize
}

func (m MP3Metadata) Album() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Album()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Album()
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

func (m MP3Metadata) Artist() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Artist()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Artist()
	}
	return ""
}

func (m MP3Metadata) AverageBitrate() int {
	// If we have a XingHeader with the Xing ID then assume VBR and calculate the
	// average bitrate
	if m.FrameData.XingHeader != nil {
		if m.FrameData.XingHeader.ID == XING_XING_ID {
			durationInSeconds := m.Duration().Seconds()
			if durationInSeconds == 0 {
				return 0
			}
			//audioDataSize in bits, check to see if we have value from Xing, if not
			//then use the approximate value
			var audioDataSize float64
			if m.FrameData.XingHeader.Bytes != nil {
				audioDataSize = float64(*m.FrameData.XingHeader.Bytes * 8)
			} else {
				audioDataSize = float64(m.approximateAudioSize() * 8)
			}
			return int(audioDataSize / durationInSeconds)
		}
	}
	//Else we assume constant bitrate
	return m.FrameHeader.Bitrate * 1000
}

func (m MP3Metadata) Comment() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Comment()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Comment()
	}
	return ""
}

func (m MP3Metadata) Composer() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Composer()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Disc() (int, int) {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Disc()
	}
	//No equivalent value for ID3v1
	return 0, 0
}

var samplesPerFrameMap = map[MPEGVersion]map[MPEGLayer]int{
	MPEGVersionReserved: map[MPEGLayer]int{
		MPEGLayerReserved: 0,
		MPEGLayer1:        0,
		MPEGLayer2:        0,
		MPEGLayer3:        0,
	},
	MPEGVersion_1: map[MPEGLayer]int{
		MPEGLayerReserved: 0,
		MPEGLayer1:        384,
		MPEGLayer2:        1152,
		MPEGLayer3:        1152,
	},
	MPEGVersion_2: map[MPEGLayer]int{
		MPEGLayerReserved: 0,
		MPEGLayer1:        384,
		MPEGLayer2:        1152,
		MPEGLayer3:        576,
	},
	MPEGVersion_2_5: map[MPEGLayer]int{
		MPEGLayerReserved: 0,
		MPEGLayer1:        384,
		MPEGLayer2:        1152,
		MPEGLayer3:        576,
	},
}

func (m MP3Metadata) Duration() time.Duration {
	if m.FrameHeader.SamplingRate <= 0 {
		return time.Duration(0)
	}
	var seconds float64
	if m.FrameData.XingHeader != nil {
		if m.FrameData.XingHeader.Frames != nil {
			spf := samplesPerFrameMap[m.FrameHeader.Version][m.FrameHeader.Layer]
			seconds = float64(*m.FrameData.XingHeader.Frames*spf) / float64(m.FrameHeader.SamplingRate)
		} else if m.FrameData.XingHeader.Bytes != nil {
			//Assume constant bitrate, use accurate byte size
			if m.FrameHeader.Bitrate <= 0 {
				return time.Duration(0)
			}
			seconds = float64(*m.FrameData.XingHeader.Bytes*8) / float64(m.FrameHeader.Bitrate*1000)
		}
	} else {
		//Assume constant bitrate, use estimated byte size
		if m.FrameHeader.Bitrate <= 0 {
			return time.Duration(0)
		}
		seconds = float64(m.approximateAudioSize()*8) / float64(m.FrameHeader.Bitrate*1000)
	}
	return time.Duration(seconds * float64(time.Second))
}

func (m MP3Metadata) Genre() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Genre()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Genre()
	}
	return ""
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

func (m MP3Metadata) Format() Format {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Format()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Format()
	}
	return UnknownFormat
}

func (m MP3Metadata) Lyrics() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Lyrics()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Picture() *Picture {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Picture()
	}
	//No equivalent value for ID3v1
	return nil
}

func (m MP3Metadata) Title() string {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Title()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Title()
	}
	return ""
}

func (m MP3Metadata) Track() (int, int) {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Track()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Track()
	}
	return 0, 0
}

func (m MP3Metadata) Year() int {
	if m.ID3v2Tags != nil {
		return m.ID3v2Tags.Year()
	} else if m.ID3v1Tags != nil {
		return m.ID3v1Tags.Year()
	}
	return 0
}
