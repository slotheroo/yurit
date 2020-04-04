// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

//VorbisComment contains tag data from a Vorbis comment. This type of tag is
//typically found in Ogg (.ogg) and FLAC (.flac) files
type VorbisComment struct {
	Size     int //Size of the comment header and comments
	Comments map[string]string
}

//ReadVorbisComment will read a Vorbis comment from an io.Reader and returns a
//pointer to a VorbisComment. The io.Reader must be positioned at the beginning
//of the Vorbis comment header to read the comment correctly.
func ReadVorbisComment(r io.Reader) (*VorbisComment, error) {
	var m = VorbisComment{Comments: make(map[string]string)}

	//Get the size of the vendor field then read it into our struct
	vendorLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}
	vendor, err := readString(r, uint(vendorLen))
	if err != nil {
		return nil, err
	}
	m.Comments["vendor"] = vendor

	//Get the length of the comments section
	commentsLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}

	//We now know that the overall size is the combined length values plus 4 bytes
	//each for the length data fields themselves
	m.Size = int(vendorLen + commentsLen + 8)

	//Iterate and read in each comment
	for i := uint32(0); i < commentsLen; i++ {
		l, err := readUint32LittleEndian(r)
		if err != nil {
			return nil, err
		}
		s, err := readString(r, uint(l))
		if err != nil {
			return nil, err
		}
		k, v, err := parseComment(s)
		if err != nil {
			return nil, err
		}
		m.Comments[strings.ToLower(k)] = v
	}
	return &m, nil
}

func parseComment(c string) (k, v string, err error) {
	kv := strings.SplitN(c, "=", 2)
	if len(kv) != 2 {
		err = errors.New("vorbis comment must contain '='")
		return
	}
	k = kv[0]
	v = kv[1]
	return
}

func (m *VorbisComment) Format() Format {
	return VORBIS
}

func (m *VorbisComment) Title() string {
	return m.Comments["title"]
}

func (m *VorbisComment) Artist() string {
	// PERFORMER
	// The artist(s) who performed the work. In classical music this would be the
	// conductor, orchestra, soloists. In an audio book it would be the actor who
	// did the reading. In popular music this is typically the same as the ARTIST
	// and is omitted.
	if m.Comments["performer"] != "" {
		return m.Comments["performer"]
	}
	return m.Comments["artist"]
}

func (m *VorbisComment) Album() string {
	return m.Comments["album"]
}

func (m *VorbisComment) AlbumArtist() string {
	// This field isn't actually included in the standard, though
	// it is commonly assigned to albumartist.
	return m.Comments["albumartist"]
}

func (m *VorbisComment) Composer() string {
	// ARTIST
	// The artist generally considered responsible for the work. In popular music
	// this is usually the performing band or singer. For classical music it would
	// be the composer. For an audio book it would be the author of the original text.
	if m.Comments["composer"] != "" {
		return m.Comments["composer"]
	}
	if m.Comments["performer"] == "" {
		return ""
	}
	return m.Comments["artist"]
}

func (m *VorbisComment) Genre() string {
	return m.Comments["genre"]
}

func (m *VorbisComment) Year() int {
	var dateFormat string

	// The date need to follow the international standard https://en.wikipedia.org/wiki/ISO_8601
	// and obviously the VorbisComment standard https://wiki.xiph.org/VorbisComment#Date_and_time
	switch len(m.Comments["date"]) {
	case 0:
		return 0
	case 4:
		dateFormat = "2006"
	case 7:
		dateFormat = "2006-01"
	case 10:
		dateFormat = "2006-01-02"
	}

	t, _ := time.Parse(dateFormat, m.Comments["date"])
	return t.Year()
}

func (m *VorbisComment) Track() (int, int) {
	x, _ := strconv.Atoi(m.Comments["tracknumber"])
	// https://wiki.xiph.org/Field_names
	n, _ := strconv.Atoi(m.Comments["tracktotal"])
	return x, n
}

func (m *VorbisComment) Disc() (int, int) {
	// https://wiki.xiph.org/Field_names
	x, _ := strconv.Atoi(m.Comments["discnumber"])
	n, _ := strconv.Atoi(m.Comments["disctotal"])
	return x, n
}

func (m *VorbisComment) Lyrics() string {
	return m.Comments["lyrics"]
}

func (m *VorbisComment) Comment() string {
	if m.Comments["comment"] != "" {
		return m.Comments["comment"]
	}
	return m.Comments["description"]
}
