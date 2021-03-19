package yurit

import (
	"fmt"
	"io"
)

// id3v2Header is a type which represents an ID3v2 tag header.
type id3v2Header struct {
	version           Format
	revision          byte
	unsynchronization bool
	extendedHeader    bool
	experimental      bool
	footer            bool
	size              int
}

// id3v2ExtendedHeader is the bytes from the extended header
type id3v2ExtendedHeader []byte

// processID3v2Header converts 10 header bytes into an id3v2Header
// If an extended header is declared in the header flags, it will be read and
// returned as well.
func processID3v2Header(b []byte, r io.ReadSeeker) (*id3v2Header, id3v2ExtendedHeader, error) {
	var h id3v2Header
	var x id3v2ExtendedHeader
	//Skip input validation. (e.g. len(b) and ID3 as first 3 bytes.
	//Assume it was taken care of in the calling function.

	versByte := b[3]
	h.revision = b[4]
	h.unsynchronization = getBit(b[5], 7)
	h.size = get7BitChunkedInt(b[6:])
	if versByte == 2 {
		h.version = ID3v2_2
		//No check here for the compression flag as it doesn't seem to be ever valid
	} else if versByte == 3 {
		h.version = ID3v2_3
		h.extendedHeader = getBit(b[5], 6)
		h.experimental = getBit(b[5], 5)
	} else if versByte == 4 {
		h.version = ID3v2_4
		h.extendedHeader = getBit(b[5], 6)
		h.experimental = getBit(b[5], 5)
		h.footer = getBit(b[5], 4)
	} else if versByte > 4 {
		return nil, nil, fmt.Errorf("Unknown ID3v2 version: %v, expected: 2, 3 or 4", uint(b[0]))
	}

	if h.extendedHeader {
		if h.version == ID3v2_3 {
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, nil, fmt.Errorf("expected to read 4 bytes (ID3v23 extended header len): %v", err)
			}
			//v2_3 size excludes the 4 size bites
			//TODO replace the below func w/ something more elegant
			extendedHeaderSize := uint(getUint32AsInt64(b))
			// skip header
			b2, err := readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, nil, fmt.Errorf("expected to read %d bytes (ID3v23 extended header): %v", extendedHeaderSize, err)
			}
			x = append(b, b2...)
		} else if h.version == ID3v2_4 {
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, nil, fmt.Errorf("expected to read 4 bytes (ID3v24 extended header len): %v", err)
			}
			//v2_4 extended header size includes the 4 size bytes themselves
			extendedHeaderSize := uint(get7BitChunkedInt(b) - 4)
			// skip header
			b2, err := readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, nil, fmt.Errorf("expected to seek %d bytes (ID3v24 extended header): %v", extendedHeaderSize, err)
			}
			x = append(b, b2...)
		}
	}

	return &h, x, nil
}
