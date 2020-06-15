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

//vorbisComment contains tag data from a Vorbis comment. This type of tag is
//typically found in Ogg (.ogg) and FLAC (.flac) files
type vorbisComment map[string]string

//readVorbisComment will read a Vorbis comment from an io.Reader and returns a
//pointer to a vorbisComment. The io.Reader must be positioned at the beginning
//of the Vorbis comment header to read the comment correctly.
func readVorbisComment(r io.Reader) (vorbisComment, error) {
	var vc = vorbisComment{}

	//Get the size of the vendor field then read it into our struct
	vendorLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}
	vendor, err := readString(r, uint(vendorLen))
	if err != nil {
		return nil, err
	}
	vc["vendor"] = vendor

	//Get the length of the comments section
	commentsLen, err := readUint32LittleEndian(r)
	if err != nil {
		return nil, err
	}

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
		key, val, err := parseComment(s)
		if err != nil {
			return nil, err
		}
		vc[strings.ToLower(key)] = val
	}
	return vc, nil
}

//processVorbisComment will process a Vorbis comment from bytes bytes
//representing a vorbis comment returns a vorbisComment
func processVorbisComment(b []byte) (vorbisComment, error) {
	var vc = vorbisComment{}

	//Get the size of the vendor field then read it into our struct
	if err := checkLen(b, 4); err != nil {
		return nil, err
	}
	vendorLen := getUint32Little(b[0:4])
	if err := checkLen(b, int(4+vendorLen)); err != nil {
		return nil, err
	}
	vc["vendor"] = string(b[4 : 4+vendorLen])

	//Get the length of the comments section
	if err := checkLen(b, int(8+vendorLen)); err != nil {
		return nil, err
	}
	commentsLen := getUint32Little(b[4+vendorLen : 8+vendorLen])
	offset := 8 + vendorLen
	//Iterate and read in each comment
	for i := uint32(0); i < commentsLen; i++ {
		if err := checkLen(b, int(offset+4)); err != nil {
			return nil, err
		}
		l := getUint32Little(b[offset : offset+4])
		if err := checkLen(b, int(offset+4+l)); err != nil {
			return nil, err
		}
		s := string(b[offset+4 : offset+4+l])
		key, val, err := parseComment(s)
		if err != nil {
			return nil, err
		}
		vc[strings.ToLower(key)] = val
		offset += 4 + l
	}
	return vc, nil
}

func parseComment(c string) (key, val string, err error) {
	kv := strings.SplitN(c, "=", 2)
	if len(kv) != 2 {
		err = errors.New("vorbis comment must contain '='")
		return
	}
	key = kv[0]
	val = kv[1]
	return
}

func (vc vorbisComment) Album() string {
	return vc["album"]
}

func (vc vorbisComment) AlbumArtist() string {
	// This field isn't actually included in the standard, though
	// it is commonly assigned to albumartist.
	return vc["albumartist"]
}

func (vc vorbisComment) Artist() string {
	// PERFORMER
	// The artist(s) who performed the work. In classical music this would be the
	// conductor, orchestra, soloists. In an audio book it would be the actor who
	// did the reading. In popular music this is typically the same as the ARTIST
	// and is omitted.
	if vc["performer"] != "" {
		return vc["performer"]
	}
	return vc["artist"]
}

func (vc vorbisComment) Comment() string {
	if vc["comment"] != "" {
		return vc["comment"]
	}
	return vc["description"]
}

func (vc vorbisComment) Composer() string {
	// ARTIST
	// The artist generally considered responsible for the work. In popular music
	// this is usually the performing band or singer. For classical music it would
	// be the composer. For an audio book it would be the author of the original text.
	if vc["composer"] != "" {
		return vc["composer"]
	}
	if vc["performer"] == "" {
		return ""
	}
	return vc["artist"]
}

func (vc vorbisComment) Disc() (int, int) {
	// https://wiki.xiph.org/Field_names
	x, _ := strconv.Atoi(vc["discnumber"])
	n, _ := strconv.Atoi(vc["disctotal"])
	return x, n
}

func (vc vorbisComment) Format() Format {
	return VORBIS
}

func (vc vorbisComment) Genre() string {
	return vc["genre"]
}

func (vc vorbisComment) Lyrics() string {
	return vc["lyrics"]
}

func (vc vorbisComment) Title() string {
	return vc["title"]
}

func (vc vorbisComment) Track() (int, int) {
	x, _ := strconv.Atoi(vc["tracknumber"])
	// https://wiki.xiph.org/Field_names
	n, _ := strconv.Atoi(vc["tracktotal"])
	return x, n
}

func (vc vorbisComment) Year() int {
	var dateFormat string

	// The date need to follow the international standard https://en.wikipedia.org/wiki/ISO_8601
	// and obviously the VorbisComment standard https://wiki.xiph.org/VorbisComment#Date_and_time
	switch len(vc["date"]) {
	case 0:
		return 0
	case 4:
		dateFormat = "2006"
	case 7:
		dateFormat = "2006-01"
	case 10:
		dateFormat = "2006-01-02"
	}

	t, _ := time.Parse(dateFormat, vc["date"])
	return t.Year()
}
