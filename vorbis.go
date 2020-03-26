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

/*func newMetadataVorbis() *metadataVorbis {
	return &metadataVorbis{
		c: make(map[string]string),
	}
}

type metadataVorbis struct {
	c map[string]string // the vorbis comments
	p *Picture
}*/

type VorbisComment struct {
	Fields map[string]string
}

func ReadVorbisComment(r io.Reader) (*VorbisComment, error) {
	var m = VorbisComment{Fields: make(map[string]string)}
	vendorLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}

	vendor, err := readString(r, uint(vendorLen))
	if err != nil {
		return nil, err
	}
	m.Fields["vendor"] = vendor

	commentsLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}

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
		m.Fields[strings.ToLower(k)] = v
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

/*func (m *VorbisComment) Raw() map[string]interface{} {
	raw := make(map[string]interface{}, len(m.Fields))
	for k, v := range m.Fields {
		raw[k] = v
	}
	return raw
}*/

func (m *VorbisComment) Title() string {
	return m.Fields["title"]
}

func (m *VorbisComment) Artist() string {
	// PERFORMER
	// The artist(s) who performed the work. In classical music this would be the
	// conductor, orchestra, soloists. In an audio book it would be the actor who
	// did the reading. In popular music this is typically the same as the ARTIST
	// and is omitted.
	if m.Fields["performer"] != "" {
		return m.Fields["performer"]
	}
	return m.Fields["artist"]
}

func (m *VorbisComment) Album() string {
	return m.Fields["album"]
}

func (m *VorbisComment) AlbumArtist() string {
	// This field isn't actually included in the standard, though
	// it is commonly assigned to albumartist.
	return m.Fields["albumartist"]
}

func (m *VorbisComment) Composer() string {
	// ARTIST
	// The artist generally considered responsible for the work. In popular music
	// this is usually the performing band or singer. For classical music it would
	// be the composer. For an audio book it would be the author of the original text.
	if m.Fields["composer"] != "" {
		return m.Fields["composer"]
	}
	if m.Fields["performer"] == "" {
		return ""
	}
	return m.Fields["artist"]
}

func (m *VorbisComment) Genre() string {
	return m.Fields["genre"]
}

func (m *VorbisComment) Year() int {
	var dateFormat string

	// The date need to follow the international standard https://en.wikipedia.org/wiki/ISO_8601
	// and obviously the VorbisComment standard https://wiki.xiph.org/VorbisComment#Date_and_time
	switch len(m.Fields["date"]) {
	case 0:
		return 0
	case 4:
		dateFormat = "2006"
	case 7:
		dateFormat = "2006-01"
	case 10:
		dateFormat = "2006-01-02"
	}

	t, _ := time.Parse(dateFormat, m.Fields["date"])
	return t.Year()
}

func (m *VorbisComment) Track() (int, int) {
	x, _ := strconv.Atoi(m.Fields["tracknumber"])
	// https://wiki.xiph.org/Field_names
	n, _ := strconv.Atoi(m.Fields["tracktotal"])
	return x, n
}

func (m *VorbisComment) Disc() (int, int) {
	// https://wiki.xiph.org/Field_names
	x, _ := strconv.Atoi(m.Fields["discnumber"])
	n, _ := strconv.Atoi(m.Fields["disctotal"])
	return x, n
}

func (m *VorbisComment) Lyrics() string {
	return m.Fields["lyrics"]
}

func (m *VorbisComment) Comment() string {
	if m.Fields["comment"] != "" {
		return m.Fields["comment"]
	}
	return m.Fields["description"]
}
