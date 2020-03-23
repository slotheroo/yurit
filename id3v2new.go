package yurit

import (
	"io"
)

type ID3v2 struct {
	Header ID3v2Header
	Frames map[string]interface{}
}

func extractID3v2(r io.ReadSeeker) (*ID3v2, error) {
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
	h, offset, err := readID3v2Header(r)
	if err != nil {
		return nil, err
	}

	var ur io.Reader = r
	if h.Unsynchronisation {
		ur = &unsynchroniser{Reader: r}
	}

	f, err := readID3v2Frames(ur, offset, h)
	if err != nil {
		return nil, err
	}

	i := ID3v2{Header: *h, Frames: f}
	return &i, nil
}
