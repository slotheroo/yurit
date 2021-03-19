package yurit

/*
import (
	"fmt"
	"io"
)

// id3v2Header is a type which represents an ID3v2 tag header.
type id3v2Header struct {
	Version           Format
	Unsynchronisation bool
	ExtendedHeader    bool
	Experimental      bool
	Footer            bool
	Size              uint
}

// readID3v2Header reads the ID3v2 header from the given io.Reader.
// offset it number of bytes of header that was read
func readID3v2Header(r io.Reader) (h *id3v2Header, offset uint, err error) {
	offset = 10
	b, err := readBytes(r, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("expected to read 10 bytes (id3v2Header): %v", err)
	}

	if string(b[0:3]) != "ID3" {
		return nil, 0, fmt.Errorf("expected to read \"ID3\"")
	}

	b = b[3:]
	var vers Format
	switch uint(b[0]) {
	case 2:
		vers = ID3v2_2
	case 3:
		vers = ID3v2_3
	case 4:
		vers = ID3v2_4
	case 0, 1:
		fallthrough
	default:
		return nil, 0, fmt.Errorf("ID3 version: %v, expected: 2, 3 or 4", uint(b[0]))
	}

	// NB: We ignore b[1] (the revision) as we don't currently rely on it.
	h = &id3v2Header{
		Version:           vers,
		Unsynchronisation: getBit(b[2], 7),
		ExtendedHeader:    getBit(b[2], 6),
		Experimental:      getBit(b[2], 5),
		Footer:            getBit(b[2], 4),
		Size:              uint(get7BitChunkedInt(b[3:7])),
	}

	if h.ExtendedHeader {
		switch vers {
		case ID3v2_3:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v23 extended header len): %v", err)
			}
			// skip header, size is excluding len bytes
			extendedHeaderSize := uint(getInt(b))
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v23 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize
		case ID3v2_4:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v24 extended header len): %v", err)
			}
			// skip header, size is synchsafe int including len bytes
			extendedHeaderSize := uint(get7BitChunkedInt(b)) - 4
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v24 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize
		default:
			// nop, only 2.3 and 2.4 should have extended header
		}
	}

	return h, offset, nil
}
*/
