// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

//id3v2Tags holds metadata from an ID3v2.2, ID3v2.3, or ID3v2.4 tag, which are
//commonly found in mp3 files.
type id3v2Tags struct {
	header         id3v2Header
	extendedHeader id3v2ExtendedHeader
	frames         map[string]interface{}
}

//ReadID3v2Tags reads ID3v2 tags from the io.ReadSeeker. If there is no ID3v2
//tag, returns nil. Method assumes that the reader has been prepositioned to
//the beginning of the tag, which should be the beginning of the file.
func ReadID3v2Tags(r io.ReadSeeker) (*id3v2Tags, error) {
	b, err := readBytes(r, 10)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			//File is not long enough to hold a tag
			return nil, nil
		}
		return nil, err
	}
	//No ID3v2 tags, reset and return nil
	if string(b[0:3]) != "ID3" {
		_, _ = r.Seek(-10, io.SeekCurrent)
		return nil, nil
	}

	h, x, err := processID3v2Header(b, r)
	if err != nil {
		return nil, err
	}

	if h.size < len(x) {
		return nil, fmt.Errorf("Could not read ID3v2 tag. Tag size of %d is less than extended header size of %d", h.size, len(x))
	}
	framesSize := uint(h.size - len(x))

	b, err = readBytes(r, framesSize)
	if err != nil {
		return nil, err
	}

	f, err := processID3v2Frames(b, h)
	if err != nil {
		return nil, err
	}

	i := id3v2Tags{header: *h, extendedHeader: x, frames: f}
	return &i, nil
}

var id3v2genreRe = regexp.MustCompile(`(.*[^(]|.* |^)\(([0-9]+)\) *(.*)$`)

var id3v2Genres = [...]string{
	"Blues", "Classic Rock", "Country", "Dance", "Disco", "Funk", "Grunge",
	"Hip-Hop", "Jazz", "Metal", "New Age", "Oldies", "Other", "Pop", "R&B",
	"Rap", "Reggae", "Rock", "Techno", "Industrial", "Alternative", "Ska",
	"Death Metal", "Pranks", "Soundtrack", "Euro-Techno", "Ambient",
	"Trip-Hop", "Vocal", "Jazz+Funk", "Fusion", "Trance", "Classical",
	"Instrumental", "Acid", "House", "Game", "Sound Clip", "Gospel",
	"Noise", "AlternRock", "Bass", "Soul", "Punk", "Space", "Meditative",
	"Instrumental Pop", "Instrumental Rock", "Ethnic", "Gothic",
	"Darkwave", "Techno-Industrial", "Electronic", "Pop-Folk",
	"Eurodance", "Dream", "Southern Rock", "Comedy", "Cult", "Gangsta",
	"Top 40", "Christian Rap", "Pop/Funk", "Jungle", "Native American",
	"Cabaret", "New Wave", "Psychedelic", "Rave", "Showtunes", "Trailer",
	"Lo-Fi", "Tribal", "Acid Punk", "Acid Jazz", "Polka", "Retro",
	"Musical", "Rock & Roll", "Hard Rock", "Folk", "Folk-Rock",
	"National Folk", "Swing", "Fast Fusion", "Bebob", "Latin", "Revival",
	"Celtic", "Bluegrass", "Avantgarde", "Gothic Rock", "Progressive Rock",
	"Psychedelic Rock", "Symphonic Rock", "Slow Rock", "Big Band",
	"Chorus", "Easy Listening", "Acoustic", "Humour", "Speech", "Chanson",
	"Opera", "Chamber Music", "Sonata", "Symphony", "Booty Bass", "Primus",
	"Porn Groove", "Satire", "Slow Jam", "Club", "Tango", "Samba",
	"Folklore", "Ballad", "Power Ballad", "Rhythmic Soul", "Freestyle",
	"Duet", "Punk Rock", "Drum Solo", "A capella", "Euro-House", "Dance Hall",
	"Goa", "Drum & Bass", "Club-House", "Hardcore", "Terror", "Indie",
	"Britpop", "Negerpunk", "Polsk Punk", "Beat", "Christian Gangsta Rap",
	"Heavy Metal", "Black Metal", "Crossover", "Contemporary Christian",
	"Christian Rock ", "Merengue", "Salsa", "Thrash Metal", "Anime", "JPop",
	"Synthpop",
}

//  id3v2genre parse a id3v2 genre tag and expand the numeric genres
func id3v2genre(genre string) string {
	c := true
	for c {
		orig := genre
		if match := id3v2genreRe.FindStringSubmatch(genre); len(match) > 0 {
			if genreID, err := strconv.Atoi(match[2]); err == nil {
				if genreID < len(id3v2Genres) {
					genre = id3v2Genres[genreID]
					if match[1] != "" {
						genre = strings.TrimSpace(match[1]) + " " + genre
					}
					if match[3] != "" {
						genre = genre + " " + match[3]
					}
				}
			}
		}
		c = (orig != genre)
	}
	return strings.Replace(genre, "((", "(", -1)
}

type frameNames map[string][2]string

func (f frameNames) Name(s string, fm Format) string {
	l, ok := f[s]
	if !ok {
		return ""
	}

	switch fm {
	case ID3v2_2:
		return l[0]
	case ID3v2_3:
		return l[1]
	case ID3v2_4:
		if s == "year" {
			return "TDRC"
		}
		return l[1]
	}
	return ""
}

var frames = frameNames(map[string][2]string{
	"title":        [2]string{"TT2", "TIT2"},
	"artist":       [2]string{"TP1", "TPE1"},
	"album":        [2]string{"TAL", "TALB"},
	"album_artist": [2]string{"TP2", "TPE2"},
	"composer":     [2]string{"TCM", "TCOM"},
	"year":         [2]string{"TYE", "TYER"},
	"track":        [2]string{"TRK", "TRCK"},
	"disc":         [2]string{"TPA", "TPOS"},
	"genre":        [2]string{"TCO", "TCON"},
	"picture":      [2]string{"PIC", "APIC"},
	"lyrics":       [2]string{"", "USLT"},
	"comment":      [2]string{"COM", "COMM"},
})

func (m id3v2Tags) getString(k string) string {
	v, ok := m.frames[k]
	if !ok {
		return ""
	}
	return v.(string)
}

func (m id3v2Tags) Format() Format              { return m.header.version }
func (m id3v2Tags) Raw() map[string]interface{} { return m.frames }

func (m id3v2Tags) Title() string {
	return m.getString(frames.Name("title", m.Format()))
}

func (m id3v2Tags) Artist() string {
	return m.getString(frames.Name("artist", m.Format()))
}

func (m id3v2Tags) Album() string {
	return m.getString(frames.Name("album", m.Format()))
}

func (m id3v2Tags) AlbumArtist() string {
	return m.getString(frames.Name("album_artist", m.Format()))
}

func (m id3v2Tags) Composer() string {
	return m.getString(frames.Name("composer", m.Format()))
}

func (m id3v2Tags) Genre() string {
	return id3v2genre(m.getString(frames.Name("genre", m.Format())))
}

func (m id3v2Tags) Year() int {
	year, _ := strconv.Atoi(m.getString(frames.Name("year", m.Format())))
	return year
}

func parseXofN(s string) (x, n int) {
	xn := strings.Split(s, "/")
	if len(xn) != 2 {
		x, _ = strconv.Atoi(s)
		return x, 0
	}
	x, _ = strconv.Atoi(strings.TrimSpace(xn[0]))
	n, _ = strconv.Atoi(strings.TrimSpace(xn[1]))
	return x, n
}

func (m id3v2Tags) Track() (int, int) {
	return parseXofN(m.getString(frames.Name("track", m.Format())))
}

func (m id3v2Tags) Disc() (int, int) {
	return parseXofN(m.getString(frames.Name("disc", m.Format())))
}

func (m id3v2Tags) Lyrics() string {
	t, ok := m.frames[frames.Name("lyrics", m.Format())]
	if !ok {
		return ""
	}
	return t.(*Comm).Text
}

func (m id3v2Tags) Comment() string {
	t, ok := m.frames[frames.Name("comment", m.Format())]
	if !ok {
		return ""
	}
	// id3v23 has Text, id3v24 has Description
	if t.(*Comm).Description == "" {
		return trimString(t.(*Comm).Text)
	}
	return trimString(t.(*Comm).Description)
}

func (m id3v2Tags) Picture() *Picture {
	v, ok := m.frames[frames.Name("picture", m.Format())]
	if !ok {
		return nil
	}
	return v.(*Picture)
}
