package yurit

import (
	"io"
	"os"
	"time"
)

//MP3Metadata is a collection of metadata from an mp3 file including tags and
//frame information.
type MP3Metadata struct {
	id3v2Tags   *id3v2Tags
	fileSize    int64
	frameHeader mpegFrameHeader
	id3v1tags   id3v1tags
	xingHeader  mp3XingHeader
}

func ReadFromMP3(file *os.File) (*MP3Metadata, error) {
	var (
		m MP3Metadata
	)
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	m.fileSize = stat.Size()
	//Extract any ID3v2 tags, if any
	id3v2, err := ReadID3v2Tags(file)
	if err != nil {
		return nil, err
	}
	m.id3v2Tags = id3v2
	//Seek to the end of the ID3v2 tags, or the beginning of the file if there are
	//no tags.
	if id3v2 != nil {
		_, err = file.Seek(int64(id3v2.header.size), io.SeekStart)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}
	//Find and read the first encountered frame header and look for xing header
	//in the data for the first frame
	err = m.readFrame(file)
	if err != nil {
		return nil, err
	}
	//Look for an ID3v1 tag at the end of the file
	id3v1, err := ReadID3v1Tags(file)
	if err != nil {
		return nil, err
	}
	m.id3v1tags = id3v1
	return &m, nil
}

func (m *MP3Metadata) readFrame(r io.ReadSeeker) error {
	frameHeader, err := readMPEGFrameHeader(r)
	if err != nil {
		return err
	}
	m.frameHeader = frameHeader
	_, err = r.Seek(int64(m.frameHeader.sideInfoLength()), io.SeekCurrent)
	if err != nil {
		return err
	}
	b, err := readBytes(r, 6)
	if err != nil {
		return err
	}
	var (
		getXing bool  = false
		goBack  int64 = -6
	)
	//Look in two possible locations for Xing/Info header id
	if string(b[0:4]) == "Xing" || string(b[0:4]) == "Info" {
		getXing = true
	} else if string(b[2:6]) == "Xing" || string(b[2:6]) == "Info" {
		getXing = true
		goBack = -4
	}
	if getXing {
		//If found go back to beginning of Xing header and grab the data
		r.Seek(goBack, io.SeekCurrent)
		xingHeader, err := readMP3XingHeader(r)
		if err != nil {
			return err
		}

		m.xingHeader = xingHeader
	}
	return err
}

// return the approximate size of audio data in bytes
func (m MP3Metadata) approximateAudioSize() int64 {
	var v1TagSize int64 = 0
	var v2TagSize int64 = 0
	if m.id3v1tags != nil {
		v1TagSize = 128
	}
	if m.id3v2Tags != nil {
		v2TagSize = 10 + int64(m.id3v2Tags.header.size)
	}
	return m.fileSize - v1TagSize - v2TagSize
}

func (m MP3Metadata) Album() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Album()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Album()
	}
	return ""
}

func (m MP3Metadata) AlbumArtist() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.AlbumArtist()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Artist() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Artist()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Artist()
	}
	return ""
}

func (m MP3Metadata) AverageBitrate() int {
	// If we have a XingHeader with the Xing ID then assume VBR and calculate the
	// average bitrate
	if m.xingHeader != nil {
		if m.xingHeader.ID() == "Xing" {
			durationInSeconds := m.Duration().Seconds()
			if durationInSeconds == 0 {
				return 0
			}
			//audioDataSize in bits, check to see if we have value from Xing, if not
			//then use the approximate value
			var audioDataSize float64
			if m.xingHeader.TotalBytes() != nil {
				audioDataSize = float64(*m.xingHeader.TotalBytes() * 8)
			} else {
				audioDataSize = float64(m.approximateAudioSize() * 8)
			}
			return int(audioDataSize / durationInSeconds)
		}
	}
	//Else we assume constant bitrate
	return m.frameHeader.Bitrate() * 1000
}

func (m MP3Metadata) Comment() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Comment()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Comment()
	}
	return ""
}

func (m MP3Metadata) Composer() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Composer()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Disc() (int, int) {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Disc()
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
	if m.frameHeader.SampleRate() <= 0 {
		return time.Duration(0)
	}
	var seconds float64
	numFrames := m.xingHeader.TotalFrames()
	//numFrames, framesOk := m.xingHeader["numberOfFrames"].(int)
	numBytes := m.xingHeader.TotalBytes()
	//numBytes, bytesOk := m.xingHeader["numberOfBytes"].(int)
	if numFrames != nil && m.frameHeader.SampleRate() > 0 {
		seconds = float64(*numFrames*m.frameHeader.SamplesPerFrame()) / float64(m.frameHeader.SampleRate())
	} else if m.frameHeader.Bitrate() <= 0 {
		return time.Duration(0)
	} else if numBytes != nil {
		seconds = float64(*numBytes*8) / float64(m.frameHeader.Bitrate()*1000)
	} else {
		seconds = float64(m.approximateAudioSize()*8) / float64(m.frameHeader.Bitrate()*1000)
	}
	return time.Duration(seconds * float64(time.Second))
}

func (m MP3Metadata) Genre() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Genre()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Genre()
	}
	return ""
}

func (m MP3Metadata) FileType() FileType {
	if m.frameHeader.Layer() == MPEGLayer3 {
		return MP3
	} else if m.frameHeader.Layer() == MPEGLayer2 {
		return MP2
	} else if m.frameHeader.Layer() == MPEGLayer1 {
		return MP1
	}
	return UnknownFileType
}

func (m MP3Metadata) Format() Format {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Format()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Format()
	}
	return UnknownFormat
}

func (m MP3Metadata) ID3v2Frames() map[string]interface{} {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.frames
	}
	return nil
}

func (m MP3Metadata) Lyrics() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Lyrics()
	}
	//No equivalent value for ID3v1
	return ""
}

func (m MP3Metadata) Picture() *Picture {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Picture()
	}
	//No equivalent value for ID3v1
	return nil
}

func (m MP3Metadata) Title() string {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Title()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Title()
	}
	return ""
}

func (m MP3Metadata) Track() (int, int) {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Track()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Track()
	}
	return 0, 0
}

func (m MP3Metadata) Year() int {
	if m.id3v2Tags != nil {
		return m.id3v2Tags.Year()
	} else if m.id3v1tags != nil {
		return m.id3v1tags.Year()
	}
	return 0
}
