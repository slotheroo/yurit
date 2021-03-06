// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yurit

/*
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
)

// DefaultUTF16WithBOMByteOrder is the byte order used when the "UTF16 with BOM" encoding
// is specified without a corresponding BOM in the data.
var DefaultUTF16WithBOMByteOrder binary.ByteOrder = binary.LittleEndian

// ID3v2.2.0 frames (see http://id3.org/id3v2-00, sec 4).
var id3v22Frames = map[string]string{
	"BUF": "Recommended buffer size",

	"CNT": "Play counter",
	"COM": "Comments",
	"CRA": "Audio encryption",
	"CRM": "Encrypted meta frame",

	"ETC": "Event timing codes",
	"EQU": "Equalization",

	"GEO": "General encapsulated object",

	"IPL": "Involved people list",

	"LNK": "Linked information",

	"MCI": "Music CD Identifier",
	"MLL": "MPEG location lookup table",

	"PIC": "Attached picture",
	"POP": "Popularimeter",

	"REV": "Reverb",
	"RVA": "Relative volume adjustment",

	"SLT": "Synchronized lyric/text",
	"STC": "Synced tempo codes",

	"TAL": "Album/Movie/Show title",
	"TBP": "BPM (Beats Per Minute)",
	"TCM": "Composer",
	"TCO": "Content type",
	"TCR": "Copyright message",
	"TDA": "Date",
	"TDY": "Playlist delay",
	"TEN": "Encoded by",
	"TFT": "File type",
	"TIM": "Time",
	"TKE": "Initial key",
	"TLA": "Language(s)",
	"TLE": "Length",
	"TMT": "Media type",
	"TOA": "Original artist(s)/performer(s)",
	"TOF": "Original filename",
	"TOL": "Original Lyricist(s)/text writer(s)",
	"TOR": "Original release year",
	"TOT": "Original album/Movie/Show title",
	"TP1": "Lead artist(s)/Lead performer(s)/Soloist(s)/Performing group",
	"TP2": "Band/Orchestra/Accompaniment",
	"TP3": "Conductor/Performer refinement",
	"TP4": "Interpreted, remixed, or otherwise modified by",
	"TPA": "Part of a set",
	"TPB": "Publisher",
	"TRC": "ISRC (International Standard Recording Code)",
	"TRD": "Recording dates",
	"TRK": "Track number/Position in set",
	"TSI": "Size",
	"TSS": "Software/hardware and settings used for encoding",
	"TT1": "Content group description",
	"TT2": "Title/Songname/Content description",
	"TT3": "Subtitle/Description refinement",
	"TXT": "Lyricist/text writer",
	"TXX": "User defined text information frame",
	"TYE": "Year",

	"UFI": "Unique file identifier",
	"ULT": "Unsychronized lyric/text transcription",

	"WAF": "Official audio file webpage",
	"WAR": "Official artist/performer webpage",
	"WAS": "Official audio source webpage",
	"WCM": "Commercial information",
	"WCP": "Copyright/Legal information",
	"WPB": "Publishers official webpage",
	"WXX": "User defined URL link frame",
}

// ID3v2.3.0 frames (see http://id3.org/id3v2.3.0#Declared_ID3v2_frames).
var id3v23Frames = map[string]string{
	"AENC": "Audio encryption]",
	"APIC": "Attached picture",
	"COMM": "Comments",
	"COMR": "Commercial frame",
	"ENCR": "Encryption method registration",
	"EQUA": "Equalization",
	"ETCO": "Event timing codes",
	"GEOB": "General encapsulated object",
	"GRID": "Group identification registration",
	"IPLS": "Involved people list",
	"LINK": "Linked information",
	"MCDI": "Music CD identifier",
	"MLLT": "MPEG location lookup table",
	"OWNE": "Ownership frame",
	"PRIV": "Private frame",
	"PCNT": "Play counter",
	"POPM": "Popularimeter",
	"POSS": "Position synchronisation frame",
	"RBUF": "Recommended buffer size",
	"RVAD": "Relative volume adjustment",
	"RVRB": "Reverb",
	"SYLT": "Synchronized lyric/text",
	"SYTC": "Synchronized tempo codes",
	"TALB": "Album/Movie/Show title",
	"TBPM": "BPM (beats per minute)",
	"TCMP": "iTunes Compilation Flag",
	"TCOM": "Composer",
	"TCON": "Content type",
	"TCOP": "Copyright message",
	"TDAT": "Date",
	"TDLY": "Playlist delay",
	"TENC": "Encoded by",
	"TEXT": "Lyricist/Text writer",
	"TFLT": "File type",
	"TIME": "Time",
	"TIT1": "Content group description",
	"TIT2": "Title/songname/content description",
	"TIT3": "Subtitle/Description refinement",
	"TKEY": "Initial key",
	"TLAN": "Language(s)",
	"TLEN": "Length",
	"TMED": "Media type",
	"TOAL": "Original album/movie/show title",
	"TOFN": "Original filename",
	"TOLY": "Original lyricist(s)/text writer(s)",
	"TOPE": "Original artist(s)/performer(s)",
	"TORY": "Original release year",
	"TOWN": "File owner/licensee",
	"TPE1": "Lead performer(s)/Soloist(s)",
	"TPE2": "Band/orchestra/accompaniment",
	"TPE3": "Conductor/performer refinement",
	"TPE4": "Interpreted, remixed, or otherwise modified by",
	"TPOS": "Part of a set",
	"TPUB": "Publisher",
	"TRCK": "Track number/Position in set",
	"TRDA": "Recording dates",
	"TRSN": "Internet radio station name",
	"TRSO": "Internet radio station owner",
	"TSIZ": "Size",
	"TSO2": "iTunes uses this for Album Artist sort order",
	"TSOC": "iTunes uses this for Composer sort order",
	"TSRC": "ISRC (international standard recording code)",
	"TSSE": "Software/Hardware and settings used for encoding",
	"TYER": "Year",
	"TXXX": "User defined text information frame",
	"UFID": "Unique file identifier",
	"USER": "Terms of use",
	"USLT": "Unsychronized lyric/text transcription",
	"WCOM": "Commercial information",
	"WCOP": "Copyright/Legal information",
	"WOAF": "Official audio file webpage",
	"WOAR": "Official artist/performer webpage",
	"WOAS": "Official audio source webpage",
	"WORS": "Official internet radio station homepage",
	"WPAY": "Payment",
	"WPUB": "Publishers official webpage",
	"WXXX": "User defined URL link frame",
}

// ID3v2.4.0 frames (see http://id3.org/id3v2.4.0-frames, sec 4).
var id3v24Frames = map[string]string{
	"AENC": "Audio encryption",
	"APIC": "Attached picture",
	"ASPI": "Audio seek point index",

	"COMM": "Comments",
	"COMR": "Commercial frame",

	"ENCR": "Encryption method registration",
	"EQU2": "Equalisation (2)",
	"ETCO": "Event timing codes",

	"GEOB": "General encapsulated object",
	"GRID": "Group identification registration",

	"LINK": "Linked information",

	"MCDI": "Music CD identifier",
	"MLLT": "MPEG location lookup table",

	"OWNE": "Ownership frame",

	"PRIV": "Private frame",
	"PCNT": "Play counter",
	"POPM": "Popularimeter",
	"POSS": "Position synchronisation frame",

	"RBUF": "Recommended buffer size",
	"RVA2": "Relative volume adjustment (2)",
	"RVRB": "Reverb",

	"SEEK": "Seek frame",
	"SIGN": "Signature frame",
	"SYLT": "Synchronised lyric/text",
	"SYTC": "Synchronised tempo codes",

	"TALB": "Album/Movie/Show title",
	"TBPM": "BPM (beats per minute)",
	"TCMP": "iTunes Compilation Flag",
	"TCOM": "Composer",
	"TCON": "Content type",
	"TCOP": "Copyright message",
	"TDEN": "Encoding time",
	"TDLY": "Playlist delay",
	"TDOR": "Original release time",
	"TDRC": "Recording time",
	"TDRL": "Release time",
	"TDTG": "Tagging time",
	"TENC": "Encoded by",
	"TEXT": "Lyricist/Text writer",
	"TFLT": "File type",
	"TIPL": "Involved people list",
	"TIT1": "Content group description",
	"TIT2": "Title/songname/content description",
	"TIT3": "Subtitle/Description refinement",
	"TKEY": "Initial key",
	"TLAN": "Language(s)",
	"TLEN": "Length",
	"TMCL": "Musician credits list",
	"TMED": "Media type",
	"TMOO": "Mood",
	"TOAL": "Original album/movie/show title",
	"TOFN": "Original filename",
	"TOLY": "Original lyricist(s)/text writer(s)",
	"TOPE": "Original artist(s)/performer(s)",
	"TOWN": "File owner/licensee",
	"TPE1": "Lead performer(s)/Soloist(s)",
	"TPE2": "Band/orchestra/accompaniment",
	"TPE3": "Conductor/performer refinement",
	"TPE4": "Interpreted, remixed, or otherwise modified by",
	"TPOS": "Part of a set",
	"TPRO": "Produced notice",
	"TPUB": "Publisher",
	"TRCK": "Track number/Position in set",
	"TRSN": "Internet radio station name",
	"TRSO": "Internet radio station owner",
	"TSO2": "iTunes uses this for Album Artist sort order",
	"TSOA": "Album sort order",
	"TSOC": "iTunes uses this for Composer sort order",
	"TSOP": "Performer sort order",
	"TSOT": "Title sort order",
	"TSRC": "ISRC (international standard recording code)",
	"TSSE": "Software/Hardware and settings used for encoding",
	"TSST": "Set subtitle",
	"TXXX": "User defined text information frame",

	"UFID": "Unique file identifier",
	"USER": "Terms of use",
	"USLT": "Unsynchronised lyric/text transcription",

	"WCOM": "Commercial information",
	"WCOP": "Copyright/Legal information",
	"WOAF": "Official audio file webpage",
	"WOAR": "Official artist/performer webpage",
	"WOAS": "Official audio source webpage",
	"WORS": "Official Internet radio station homepage",
	"WPAY": "Payment",
	"WPUB": "Publishers official webpage",
	"WXXX": "User defined URL link frame",
}

// ID3 frames that are defined in the specs.
var id3Frames = map[Format]map[string]string{
	ID3v2_2: id3v22Frames,
	ID3v2_3: id3v23Frames,
	ID3v2_4: id3v24Frames,
}

func validID3Frame(version Format, name string) bool {
	names, ok := id3Frames[version]
	if !ok {
		return false
	}
	_, ok = names[name]
	return ok
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

//TODO READ ENTIRE TAG THEN PROCESS FRAMES

// readID3v2Frames reads ID3v2 frames from the given reader using the id3v2Header.
func readID3v2Frames(r io.Reader, offset uint, h *id3v2Header) (map[string]interface{}, error) {
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

func readWFrame(b []byte) (string, error) {
	// Frame text is always encoded in ISO-8859-1
	b = append([]byte{0}, b...)
	return readTFrame(b)
}

func readTFrame(b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}

	txt, err := decodeText(b[0], b[1:])
	if err != nil {
		return "", err
	}
	return strings.Join(strings.Split(txt, string(singleZero)), ""), nil
}

const (
	encodingISO8859      byte = 0
	encodingUTF16WithBOM byte = 1
	encodingUTF16        byte = 2
	encodingUTF8         byte = 3
)

func decodeText(enc byte, b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}

	switch enc {
	case encodingISO8859: // ISO-8859-1
		return decodeISO8859(b), nil

	case encodingUTF16WithBOM: // UTF-16 with byte order marker
		if len(b) == 1 {
			return "", nil
		}
		return decodeUTF16WithBOM(b)

	case encodingUTF16: // UTF-16 without byte order (assuming BigEndian)
		if len(b) == 1 {
			return "", nil
		}
		return decodeUTF16(b, binary.BigEndian)

	case encodingUTF8: // UTF-8
		return string(b), nil

	default: // Fallback to ISO-8859-1
		return decodeISO8859(b), nil
	}
}

var (
	singleZero = []byte{0}
	doubleZero = []byte{0, 0}
)

func dataSplit(b []byte, enc byte) [][]byte {
	delim := singleZero
	if enc == encodingUTF16 || enc == encodingUTF16WithBOM {
		delim = doubleZero
	}

	result := bytes.SplitN(b, delim, 2)
	if len(result) != 2 {
		return result
	}

	if len(result[1]) == 0 {
		return result
	}

	if result[1][0] == 0 {
		// there was a double (or triple) 0 and we cut too early
		result[0] = append(result[0], result[1][0])
		result[1] = result[1][1:]
	}
	return result
}

func decodeISO8859(b []byte) string {
	r := make([]rune, len(b))
	for i, x := range b {
		r[i] = rune(x)
	}
	return string(r)
}

func decodeUTF16WithBOM(b []byte) (string, error) {
	if len(b) < 2 {
		return "", errors.New("invalid encoding: expected at least 2 bytes for UTF-16 byte order mark")
	}

	var bo binary.ByteOrder
	switch {
	case b[0] == 0xFE && b[1] == 0xFF:
		bo = binary.BigEndian
		b = b[2:]

	case b[0] == 0xFF && b[1] == 0xFE:
		bo = binary.LittleEndian
		b = b[2:]

	default:
		bo = DefaultUTF16WithBOMByteOrder
	}
	return decodeUTF16(b, bo)
}

func decodeUTF16(b []byte, bo binary.ByteOrder) (string, error) {
	if len(b)%2 != 0 {
		return "", errors.New("invalid encoding: expected even number of bytes for UTF-16 encoded text")
	}
	s := make([]uint16, 0, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		s = append(s, bo.Uint16(b[i:i+2]))
	}
	return string(utf16.Decode(s)), nil
}

// Comm is a type used in COMM, UFID, TXXX, WXXX and USLT tag.
// It's a text with a description and a specified language
// For WXXX, TXXX and UFID, we don't set a Language
type Comm struct {
	Language    string
	Description string
	Text        string
}

// String returns a string representation of the underlying Comm instance.
func (t Comm) String() string {
	if t.Language != "" {
		return fmt.Sprintf("Text{Lang: '%v', Description: '%v', %v lines}",
			t.Language, t.Description, strings.Count(t.Text, "\n"))
	}
	return fmt.Sprintf("Text{Description: '%v', %v}", t.Description, t.Text)
}

// IDv2.{3,4}
// -- Header
// <Header for 'Unsynchronised lyrics/text transcription', ID: "USLT">
// <Header for 'Comment', ID: "COMM">
// -- readTextWithDescrFrame(data, true, true)
// Text encoding       $xx
// Language            $xx xx xx
// Content descriptor  <text string according to encoding> $00 (00)
// Lyrics/text         <full text string according to encoding>
// -- Header
// <Header for         'User defined text information frame', ID: "TXXX">
// <Header for         'User defined URL link frame', ID: "WXXX">
// -- readTextWithDescrFrame(data, false, <isDataEncoded>)
// Text encoding       $xx
// Description         <text string according to encoding> $00 (00)
// Value               <text string according to encoding>
func readTextWithDescrFrame(b []byte, hasLang bool, encoded bool) (*Comm, error) {
	enc := b[0]
	b = b[1:]

	c := &Comm{}
	if hasLang {
		c.Language = string(b[:3])
		b = b[3:]
	}

	descTextSplit := dataSplit(b, enc)
	if len(descTextSplit) < 1 {
		return nil, fmt.Errorf("error decoding tag description text: invalid encoding")
	}

	desc, err := decodeText(enc, descTextSplit[0])
	if err != nil {
		return nil, fmt.Errorf("error decoding tag description text: %v", err)
	}
	c.Description = desc

	if len(descTextSplit) == 1 {
		return c, nil
	}

	if !encoded {
		enc = byte(0)
	}
	text, err := decodeText(enc, descTextSplit[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding tag text: %v", err)
	}
	c.Text = text

	return c, nil
}

// UFID is composed of a provider (frequently a URL and a binary identifier)
// The identifier can be a text (Musicbrainz use texts, but not necessary)
type UFID struct {
	Provider   string
	Identifier []byte
}

func (u UFID) String() string {
	return fmt.Sprintf("%v (%v)", u.Provider, string(u.Identifier))
}

func readUFID(b []byte) (*UFID, error) {
	result := bytes.SplitN(b, singleZero, 2)
	if len(result) != 2 {
		return nil, errors.New("expected to split UFID data into 2 pieces")
	}

	return &UFID{
		Provider:   string(result[0]),
		Identifier: result[1],
	}, nil
}

var pictureTypes = map[byte]string{
	0x00: "Other",
	0x01: "32x32 pixels 'file icon' (PNG only)",
	0x02: "Other file icon",
	0x03: "Cover (front)",
	0x04: "Cover (back)",
	0x05: "Leaflet page",
	0x06: "Media (e.g. lable side of CD)",
	0x07: "Lead artist/lead performer/soloist",
	0x08: "Artist/performer",
	0x09: "Conductor",
	0x0A: "Band/Orchestra",
	0x0B: "Composer",
	0x0C: "Lyricist/text writer",
	0x0D: "Recording Location",
	0x0E: "During recording",
	0x0F: "During performance",
	0x10: "Movie/video screen capture",
	0x11: "A bright coloured fish",
	0x12: "Illustration",
	0x13: "Band/artist logotype",
	0x14: "Publisher/Studio logotype",
}

// Picture is a type which represents an attached picture extracted from metadata.
type Picture struct {
	Ext         string // Extension of the picture file.
	MIMEType    string // MIMEType of the picture.
	Type        string // Type of the picture (see pictureTypes).
	Description string // Description.
	Data        []byte // Raw picture data.
}

// String returns a string representation of the underlying Picture instance.
func (p Picture) String() string {
	return fmt.Sprintf("Picture{Ext: %v, MIMEType: %v, Type: %v, Description: %v, Data.Size: %v}",
		p.Ext, p.MIMEType, p.Type, p.Description, len(p.Data))
}

// IDv2.2
// -- Header
// Attached picture   "PIC"
// Frame size         $xx xx xx
// -- readPICFrame
// Text encoding      $xx
// Image format       $xx xx xx
// Picture type       $xx
// Description        <textstring> $00 (00)
// Picture data       <binary data>
func readPICFrame(b []byte) (*Picture, error) {
	enc := b[0]
	ext := string(b[1:4])
	picType := b[4]

	descDataSplit := dataSplit(b[5:], enc)
	if len(descDataSplit) != 2 {
		return nil, errors.New("error decoding PIC description text: invalid encoding")
	}
	desc, err := decodeText(enc, descDataSplit[0])
	if err != nil {
		return nil, fmt.Errorf("error decoding PIC description text: %v", err)
	}

	var mimeType string
	switch ext {
	case "jpeg", "jpg":
		mimeType = "image/jpeg"
	case "png":
		mimeType = "image/png"
	}

	return &Picture{
		Ext:         ext,
		MIMEType:    mimeType,
		Type:        pictureTypes[picType],
		Description: desc,
		Data:        descDataSplit[1],
	}, nil
}

// IDv2.{3,4}
// -- Header
// <Header for 'Attached picture', ID: "APIC">
// -- readAPICFrame
// Text encoding   $xx
// MIME type       <text string> $00
// Picture type    $xx
// Description     <text string according to encoding> $00 (00)
// Picture data    <binary data>
func readAPICFrame(b []byte) (*Picture, error) {
	enc := b[0]
	mimeDataSplit := bytes.SplitN(b[1:], singleZero, 2)
	mimeType := string(mimeDataSplit[0])

	b = mimeDataSplit[1]
	if len(b) < 1 {
		return nil, fmt.Errorf("error decoding APIC mimetype")
	}
	picType := b[0]

	descDataSplit := dataSplit(b[1:], enc)
	if len(descDataSplit) != 2 {
		return nil, errors.New("error decoding APIC description text: invalid encoding")
	}
	desc, err := decodeText(enc, descDataSplit[0])
	if err != nil {
		return nil, fmt.Errorf("error decoding APIC description text: %v", err)
	}

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	}

	return &Picture{
		Ext:         ext,
		MIMEType:    mimeType,
		Type:        pictureTypes[picType],
		Description: desc,
		Data:        descDataSplit[1],
	}, nil
}
*/
