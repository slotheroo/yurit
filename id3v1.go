package yurit

import (
	"io"
	"strconv"
)

//ID3v1Tags holds metadata from an ID3v1 (or ID3v1.1) tag, which is sometimes
//found at the end of an mp3 file. They may even be found when an ID3v2 tag is
//included separately in the file.
//http://id3.org/ID3v1
type ID3v1Tags struct {
	Frames map[string]interface{}
}

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
func ReadID3v1Tags(r io.ReadSeeker) (*ID3v1Tags, error) {
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

	f := make(map[string]interface{})

	f["title"] = getString(b[3:33])
	f["artist"] = getString(b[33:63])
	f["album"] = getString(b[63:93])
	f["year"] = getString(b[93:97])

	if b[125] == 0 {
		f["comment"] = getString(b[97:125])
		f["track"] = int(b[126])
	} else {
		f["comment"] = getString(b[97:127])
		f["track"] = 0
	}

	genreID := int(b[127])
	if genreID < len(id3v1Genres) {
		f["genre"] = id3v1Genres[genreID]
	} else {
		f["genre"] = ""
	}

	m := ID3v1Tags{Frames: f}
	return &m, nil
}

func (ID3v1Tags) Format() Format                { return ID3v1 }
func (m ID3v1Tags) Raw() map[string]interface{} { return m.Frames }

func (m ID3v1Tags) Title() string  { return m.Frames["title"].(string) }
func (m ID3v1Tags) Album() string  { return m.Frames["album"].(string) }
func (m ID3v1Tags) Artist() string { return m.Frames["artist"].(string) }
func (m ID3v1Tags) Genre() string  { return m.Frames["genre"].(string) }

func (m ID3v1Tags) Year() int {
	y := m.Frames["year"].(string)
	n, err := strconv.Atoi(y)
	if err != nil {
		return 0
	}
	return n
}

func (m ID3v1Tags) Track() (int, int) { return m.Frames["track"].(int), 0 }

/*func (m ID3v1Tags) AlbumArtist() string { return "" }
func (m ID3v1Tags) Composer() string    { return "" }
func (ID3v1Tags) Disc() (int, int)      { return 0, 0 }
func (m ID3v1Tags) Picture() *Picture   { return nil }
func (m ID3v1Tags) Lyrics() string      { return "" }*/
func (m ID3v1Tags) Comment() string { return m.Frames["comment"].(string) }
