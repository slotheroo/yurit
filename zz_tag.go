// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tag provides MP3 (ID3: v1, 2.2, 2.3 and 2.4), MP4, FLAC and OGG metadata detection,
// parsing and artwork extraction.
//
// Detect and parse tag metadata from an io.ReadSeeker (i.e. an *os.File):
// 	m, err := tag.ReadFrom(f)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Print(m.Format()) // The detected format.
// 	log.Print(m.Title())  // The title of the track (see Metadata interface for more details).
package yurit

/*
// ErrNoTagsFound is the error returned by ReadFrom when the metadata format
// cannot be identified.
var ErrNoTagsFound = errors.New("no tags found")

// ReadFrom detects and parses audio file metadata tags (currently supports ID3v1,2.{2,3,4}, MP4, FLAC/OGG).
// Returns non-nil error if the format of the given data could not be determined, or if there was a problem
// parsing the data.
func ReadFrom(r io.ReadSeeker) (Metadata, error) {
	b, err := readBytes(r, 11)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(-11, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("could not seek back to original position: %v", err)
	}

	switch {
	case string(b[0:4]) == "fLaC":
		return ReadFLACTags(r)

	case string(b[0:4]) == "OggS":
		return ReadOGGTags(r)

	case string(b[4:8]) == "ftyp":
		return ReadAtoms(r)

	case string(b[0:3]) == "ID3":
		return ReadID3v2Tags(r)

	case string(b[0:4]) == "DSD ":
		return ReadDSFTags(r)
	}

	m, err := ReadID3v1Tags(r)
	if err != nil {
		if err == ErrNotID3v1 {
			err = ErrNoTagsFound
		}
		return nil, err
	}
	return m, nil
}*/
