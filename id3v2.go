// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

/*
import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

//ID3v2Tags holds metadata from an ID3v2.2, ID3v2.3, or ID3v2.4 tag, which are
//commonly found in mp3 files.
type ID3v2Tags struct {
	header id3v2header
	Frames map[string]interface{}
}

type id3v2header map[string]interface{}

//ReadID3v2Tags reads ID3v2 tags from the io.ReadSeeker. If there is no ID3v2
//tag, returns nil. Method assumes that the reader has been preopositioned to
//the beginning of the tag, which should be the beginning of the file.
func ReadID3v2Tags(r io.ReadSeeker) (*ID3v2Tags, error) {
	b, err := readBytes(r, 3)
	if err != nil {
		return nil, err
	}
	//No ID3v2 tags, return nil
	if string(b) != "ID3" {
		return nil, nil
	}
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	h, err := readID3v2Header(r)
	if err != nil {
		return nil, err
	}

	var ur io.Reader = r
	if h.Unsynchronisation() {
		ur = &unsynchroniser{Reader: r}
	}

	f, err := readID3v2Frames(ur, offset, h)
	if err != nil {
		return nil, err
	}

	i := ID3v2Tags{Header: *h, Frames: f}
	return &i, nil
}

// ID3v2Header is a type which represents an ID3v2 tag header.
type ID3v2Header struct {
	Version           Format
	Unsynchronisation bool
	ExtendedHeader    bool
	Experimental      bool
	Footer            bool
	Size              uint
}

// readID3v2Header reads the ID3v2 header from the given io.Reader.
// offset it number of bytes of header that was read
func readID3v2Header(r io.Reader) (id3v2header, error) {
	b, err := readBytes(r, 10)
	if err != nil {
		return nil, fmt.Errorf("expected to read 10 bytes (ID3v2Header): %v", err)
	}

	if string(b[0:3]) != "ID3" {
		return nil, fmt.Errorf("expected to read \"ID3\"")
	}

	m := id3v2header{}

	if b[3] > 4 {
		return nil, fmt.Errorf("Unknown ID3 version: %v. Will not attempt to process.", b[3])
	}
	m[VersionKey] = b[3]
	format := m.Format()

	m[RevisionKey] = b[4]
	m["unsynchronisationFlag"] = getBit(b[5], 7)

	var extendedHeader bool
	byte5bit6 := getBit(b[5], 6)
	if format == ID3v2_2 {
		m[CompressionKey] = byte5bit6
	} else {
		extendedHeader = byte5bit6
		m["extendedHeaderFlag"] = extendedHeader
	}

	m["experimentalFlag"] = getBit(b[5], 5)
	m["footerFlag"] = getBit(b[5], 4)
	m["size"] = get7BitChunkedInt(b[6:10])

	if extendedHeader {
		switch format {
		case ID3v2_3:
			b, err = readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v23 extended header len): %v", err)
			}
			// skip header, size is excluding len bytes
			extendedHeaderSize := uint(getInt(b))
			b, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v23 skip extended header): %v", extendedHeaderSize, err)
			}
			m["extendedHeader"] = b
		case ID3v2_4:
			b, err = readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v24 extended header len): %v", err)
			}
			// skip header, size is synchsafe int including len bytes
			extendedHeaderSize := uint(get7BitChunkedInt(b)) - 4
			b, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v24 skip extended header): %v", extendedHeaderSize, err)
			}
			m["extendedHeader"] = b
		default:
			// nop, only 2.3 and 2.4 should have extended header
		}
	}
	return m, nil
}

func (m id3v2header) Format() Format {
	v, ok := m[VersionKey].(byte)
	if !ok {
		return UnknownFormat
	}
	if v == 2 {
		return ID3v2_2
	} else if v == 3 {
		return ID3v2_3
	} else if v == 4 {
		return ID3v2_4
	}
	return UnknownFormat
}

func (m id3v2header) Unsynchronisation() bool {
	v, _ := m["unsynchronisationFlag"].(bool)
	return v
}

// id3v2FrameFlags is a type which represents the flags which can be set on an ID3v2 frame.
type id3v2FrameFlags struct {
	// Message (ID3 2.3.0 and 2.4.0)
	TagAlterPreservation  bool
	FileAlterPreservation bool
	ReadOnly              bool

	// Format (ID3 2.3.0 and 2.4.0)
	Compression   bool
	Encryption    bool
	GroupIdentity bool
	// ID3 2.4.0 only (see http://id3.org/id3v2.4.0-structure sec 4.1)
	Unsynchronisation   bool
	DataLengthIndicator bool
}

func readID3v23FrameFlags(r io.Reader) (*id3v2FrameFlags, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return nil, err
	}

	msg := b[0]
	fmt := b[1]

	return &id3v2FrameFlags{
		TagAlterPreservation:  getBit(msg, 7),
		FileAlterPreservation: getBit(msg, 6),
		ReadOnly:              getBit(msg, 5),
		Compression:           getBit(fmt, 7),
		Encryption:            getBit(fmt, 6),
		GroupIdentity:         getBit(fmt, 5),
	}, nil
}

func readID3v24FrameFlags(r io.Reader) (*id3v2FrameFlags, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return nil, err
	}

	msg := b[0]
	fmt := b[1]

	return &id3v2FrameFlags{
		TagAlterPreservation:  getBit(msg, 6),
		FileAlterPreservation: getBit(msg, 5),
		ReadOnly:              getBit(msg, 4),
		GroupIdentity:         getBit(fmt, 6),
		Compression:           getBit(fmt, 3),
		Encryption:            getBit(fmt, 2),
		Unsynchronisation:     getBit(fmt, 1),
		DataLengthIndicator:   getBit(fmt, 0),
	}, nil

}

func readID3v2_2FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 3)
	if err != nil {
		return
	}
	size, err = readUint(r, 3)
	if err != nil {
		return
	}
	headerSize = 6
	return
}

func readID3v2_3FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 4)
	if err != nil {
		return
	}
	size, err = readUint(r, 4)
	if err != nil {
		return
	}
	headerSize = 8
	return
}

func readID3v2_4FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 4)
	if err != nil {
		return
	}
	size, err = read7BitChunkedUint(r, 4)
	if err != nil {
		return
	}
	headerSize = 8
	return
}

// readID3v2Frames reads ID3v2 frames from the given reader using the ID3v2Header.
func readID3v2Frames(r io.Reader, offset uint, h *ID3v2Header) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for offset < h.Size {
		var err error
		var name string
		var size, headerSize uint
		var flags *id3v2FrameFlags

		switch h.Version {
		case ID3v2_2:
			name, size, headerSize, err = readID3v2_2FrameHeader(r)

		case ID3v2_3:
			name, size, headerSize, err = readID3v2_3FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v23FrameFlags(r)
			headerSize += 2

		case ID3v2_4:
			name, size, headerSize, err = readID3v2_4FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v24FrameFlags(r)
			headerSize += 2
		}

		if err != nil {
			return nil, err
		}

		// FIXME: Do we still need this?
		// if size=0, we certainly are in a padding zone. ignore the rest of
		// the tags
		if size == 0 {
			break
		}

		offset += headerSize + size

		// Avoid corrupted padding (see http://id3.org/Compliance%20Issues).
		if !validID3Frame(h.Version, name) && offset > h.Size {
			break
		}

		if flags != nil {
			if flags.Compression {
				_, err = read7BitChunkedUint(r, 4) // read 4
				if err != nil {
					return nil, err
				}
				size -= 4
			}

			if flags.Encryption {
				_, err = readBytes(r, 1) // read 1 byte of encryption method
				if err != nil {
					return nil, err
				}
				size--
			}
		}

		b, err := readBytes(r, size)
		if err != nil {
			return nil, err
		}

		// There can be multiple tag with the same name. Append a number to the
		// name if there is more than one.
		rawName := name
		if _, ok := result[rawName]; ok {
			for i := 0; ok; i++ {
				rawName = name + "_" + strconv.Itoa(i)
				_, ok = result[rawName]
			}
		}

		switch {
		case name == "TXXX" || name == "TXX":
			t, err := readTextWithDescrFrame(b, false, true) // no lang, but enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name[0] == 'T':
			txt, err := readTFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = txt

		case name == "UFID" || name == "UFI":
			t, err := readUFID(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name == "WXXX" || name == "WXX":
			t, err := readTextWithDescrFrame(b, false, false) // no lang, no enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name[0] == 'W':
			txt, err := readWFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = txt

		case name == "COMM" || name == "COM" || name == "USLT" || name == "ULT":
			t, err := readTextWithDescrFrame(b, true, true) // both lang and enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name == "APIC":
			p, err := readAPICFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = p

		case name == "PIC":
			p, err := readPICFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = p

		default:
			result[rawName] = b
		}
	}
	return result, nil
}

// readID3v2Frames reads ID3v2 frames from the given reader using the ID3v2Header.
func readID3v2Frames(r io.Reader, h id3v2header) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for offset < h.Size {
		var err error
		var name string
		var size, headerSize uint
		var flags *id3v2FrameFlags

		switch h.Version {
		case ID3v2_2:
			name, size, headerSize, err = readID3v2_2FrameHeader(r)

		case ID3v2_3:
			name, size, headerSize, err = readID3v2_3FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v23FrameFlags(r)
			headerSize += 2

		case ID3v2_4:
			name, size, headerSize, err = readID3v2_4FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v24FrameFlags(r)
			headerSize += 2
		}

		if err != nil {
			return nil, err
		}

		// FIXME: Do we still need this?
		// if size=0, we certainly are in a padding zone. ignore the rest of
		// the tags
		if size == 0 {
			break
		}

		offset += headerSize + size

		// Avoid corrupted padding (see http://id3.org/Compliance%20Issues).
		if !validID3Frame(h.Version, name) && offset > h.Size {
			break
		}

		if flags != nil {
			if flags.Compression {
				_, err = read7BitChunkedUint(r, 4) // read 4
				if err != nil {
					return nil, err
				}
				size -= 4
			}

			if flags.Encryption {
				_, err = readBytes(r, 1) // read 1 byte of encryption method
				if err != nil {
					return nil, err
				}
				size--
			}
		}

		b, err := readBytes(r, size)
		if err != nil {
			return nil, err
		}

		// There can be multiple tag with the same name. Append a number to the
		// name if there is more than one.
		rawName := name
		if _, ok := result[rawName]; ok {
			for i := 0; ok; i++ {
				rawName = name + "_" + strconv.Itoa(i)
				_, ok = result[rawName]
			}
		}

		switch {
		case name == "TXXX" || name == "TXX":
			t, err := readTextWithDescrFrame(b, false, true) // no lang, but enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name[0] == 'T':
			txt, err := readTFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = txt

		case name == "UFID" || name == "UFI":
			t, err := readUFID(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name == "WXXX" || name == "WXX":
			t, err := readTextWithDescrFrame(b, false, false) // no lang, no enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name[0] == 'W':
			txt, err := readWFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = txt

		case name == "COMM" || name == "COM" || name == "USLT" || name == "ULT":
			t, err := readTextWithDescrFrame(b, true, true) // both lang and enc
			if err != nil {
				return nil, err
			}
			result[rawName] = t

		case name == "APIC":
			p, err := readAPICFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = p

		case name == "PIC":
			p, err := readPICFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = p

		default:
			result[rawName] = b
		}
	}
	return result, nil
}

type unsynchroniser struct {
	io.Reader
	ff bool
}

// filter io.Reader which skip the Unsynchronisation bytes
func (r *unsynchroniser) Read(p []byte) (int, error) {
	b := make([]byte, 1)
	i := 0
	for i < len(p) {
		if n, err := r.Reader.Read(b); err != nil || n == 0 {
			return i, err
		}
		if r.ff && b[0] == 0x00 {
			r.ff = false
			continue
		}
		p[i] = b[0]
		i++
		r.ff = (b[0] == 0xFF)
	}
	return i, nil
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

func (m ID3v2Tags) getString(k string) string {
	v, ok := m.Frames[k]
	if !ok {
		return ""
	}
	return v.(string)
}

func (m ID3v2Tags) Format() Format              { return m.Header.Version }
func (m ID3v2Tags) Raw() map[string]interface{} { return m.Frames }

func (m ID3v2Tags) Title() string {
	return m.getString(frames.Name("title", m.Format()))
}

func (m ID3v2Tags) Artist() string {
	return m.getString(frames.Name("artist", m.Format()))
}

func (m ID3v2Tags) Album() string {
	return m.getString(frames.Name("album", m.Format()))
}

func (m ID3v2Tags) AlbumArtist() string {
	return m.getString(frames.Name("album_artist", m.Format()))
}

func (m ID3v2Tags) Composer() string {
	return m.getString(frames.Name("composer", m.Format()))
}

func (m ID3v2Tags) Genre() string {
	return id3v2genre(m.getString(frames.Name("genre", m.Format())))
}

func (m ID3v2Tags) Year() int {
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

func (m ID3v2Tags) Track() (int, int) {
	return parseXofN(m.getString(frames.Name("track", m.Format())))
}

func (m ID3v2Tags) Disc() (int, int) {
	return parseXofN(m.getString(frames.Name("disc", m.Format())))
}

func (m ID3v2Tags) Lyrics() string {
	t, ok := m.Frames[frames.Name("lyrics", m.Format())]
	if !ok {
		return ""
	}
	return t.(*Comm).Text
}

func (m ID3v2Tags) Comment() string {
	t, ok := m.Frames[frames.Name("comment", m.Format())]
	if !ok {
		return ""
	}
	// id3v23 has Text, id3v24 has Description
	if t.(*Comm).Description == "" {
		return trimString(t.(*Comm).Text)
	}
	return trimString(t.(*Comm).Description)
}

func (m ID3v2Tags) Picture() *Picture {
	v, ok := m.Frames[frames.Name("picture", m.Format())]
	if !ok {
		return nil
	}
	return v.(*Picture)
}
*/
