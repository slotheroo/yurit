package yurit

import (
	"io"
	"strconv"
)

//id3v1tags holds metadata from an ID3v1 (or ID3v1.1) tag, which is sometimes
//found at the end of an mp3 file. They may even be found when an ID3v2 tag is
//included separately in the file.
//http://id3.org/ID3v1
type id3v1tags map[string]interface{}

// id3v1Genres is a list of genres as given in the ID3v1 specification.
var id3v1Genres = [...]string{
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
	"Cabaret", "New Wave", "Psychadelic", "Rave", "Showtunes", "Trailer",
	"Lo-Fi", "Tribal", "Acid Punk", "Acid Jazz", "Polka", "Retro",
	"Musical", "Rock & Roll", "Hard Rock", "Folk", "Folk-Rock",
	"National Folk", "Swing", "Fast Fusion", "Bebob", "Latin", "Revival",
	"Celtic", "Bluegrass", "Avantgarde", "Gothic Rock", "Progressive Rock",
	"Psychedelic Rock", "Symphonic Rock", "Slow Rock", "Big Band",
	"Chorus", "Easy Listening", "Acoustic", "Humour", "Speech", "Chanson",
	"Opera", "Chamber Music", "Sonata", "Symphony", "Booty Bass", "Primus",
	"Porn Groove", "Satire", "Slow Jam", "Club", "Tango", "Samba",
	"Folklore", "Ballad", "Power Ballad", "Rhythmic Soul", "Freestyle",
	"Duet", "Punk Rock", "Drum Solo", "Acapella", "Euro-House", "Dance Hall",
}

//ReadID3v1Tags reads ID3v1 tags from the io.ReadSeeker. If there is no ID3v1
//tag, returns nil.
func ReadID3v1Tags(r io.ReadSeeker) (id3v1tags, error) {
	//And ID3v1 tag will always be 128 bytes from the end of the file
	_, err := r.Seek(-128, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	b, err := readBytes(r, 128)
	if err != nil {
		return nil, err
	}

	//If this doesn't match then we don't have an ID3v1 tag
	tag := getString(b[0:3])
	if tag != "TAG" {
		return nil, nil
	}

	m := id3v1tags{}

	m["title"] = getString(b[3:33])
	m["artist"] = getString(b[33:63])
	m["album"] = getString(b[63:93])
	m["year"] = getString(b[93:97])

	if b[125] == 0 {
		m["comment"] = getString(b[97:125])
		m["track"] = int(b[126])
	} else {
		m["comment"] = getString(b[97:127])
		m["track"] = 0
	}

	genreID := int(b[127])
	if genreID < len(id3v1Genres) {
		m["genre"] = id3v1Genres[genreID]
	} else {
		m["genre"] = ""
	}

	return m, nil
}

func (m id3v1tags) Album() string {
	s, _ := m["album"].(string)
	return s
}

func (m id3v1tags) Artist() string {
	s, _ := m["artist"].(string)
	return s
}

func (m id3v1tags) Format() Format {
	return ID3v1
}

func (m id3v1tags) Raw() map[string]interface{} {
	return m
}

func (m id3v1tags) Title() string {
	s, _ := m["title"].(string)
	return s
}

func (m id3v1tags) Genre() string {
	s, _ := m["genre"].(string)
	return s
}

func (m id3v1tags) Year() int {
	s, _ := m["year"].(string)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func (m id3v1tags) Track() (int, int) {
	n, _ := m["track"].(int)
	return n, 0
}

func (m id3v1tags) Comment() string {
	s, _ := m["comment"].(string)
	return s
}
