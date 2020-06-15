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
	"time"
)

//ESDSAtom contains select information from the elementary stream descriptor
//atom as per ISO/IEC 14496-1.
type ESDSAtom struct {
	ObjectProfileIndication byte //0x40 == Audio 14496-3 AAC Main
	StreamType              byte //0x05 == AudioStream
	MaxBitrate              int
	AvgBitrate              int //average bitrate
}

//MovieHeaderAtom contains select information from the Movie Header (mvhd) atom.
//This includes the duration and time scale. These two fiels are related. The
//time scale represents how many units occur per second in the stream -- which
//for audio means sample rate. The duration represents the total number of those
//units in the stream.
type MovieHeaderAtom struct {
	TimeScale int
	Duration  int
}

//MP4AAtom contains select information from the mp4a Sample Descriptor atom.
//The specific data extracted include the number of channels and the sample
//rate. The sample rate here should match the time scale in the MovieHeaderAtom.
type MP4AAtom struct {
	Version    int
	Channels   int
	SampleRate float64
}

// MP4Metadata is the implementation of Metadata for MP4 tag (atom) data.
type MP4Metadata struct {
	fileType        FileType
	data            map[string]interface{}
	movieHeaderAtom *MovieHeaderAtom
	mp4aAtom        *MP4AAtom
	esdsAtom        *ESDSAtom
}

// ReadMP4 reads MP4 metadata atoms from the io.ReadSeeker into a Metadata, returning
// non-nil error if there was a problem.
func ReadMP4(r io.ReadSeeker) (MP4Metadata, error) {
	m := MP4Metadata{
		data:     make(map[string]interface{}),
		fileType: UnknownFileType,
	}
	err := m.readAtoms(r)
	return m, err
}

// NB: atomMetaItemNames does not include "----", this is handled separately
var atomMetaItemNames = []string{
	"\xa9alb", //album
	"\xa9art", //artist
	"\xa9ART", //artist
	"aART",    //album artist
	"\xa9day", //date
	"\xa9nam", //title
	"\xa9gen", //genre
	"trkn",    //track number
	"\xa9wrt", //composer
	"\xa9too", //encoder
	"cprt",    //copyright
	"covr",    //picture
	"\xa9grp", //grouping
	"keyw",    //keyword
	"\xa9lyr", //lyrics
	"\xa9cmt", //comment
	"tmpo",    //tempo
	"cpil",    //compilation
	"disk",    //disk number
}

//Atoms where we want to look for child atoms and how many bytes to skip when
//moving to the child level. (Some atoms have version, flags etc. after the name
//but before the child atom header.)
var parentAtomNamesSkipBytes = map[string]int64{
	"ilst": 0,
	"mdia": 0,
	"meta": 4, //1 byte version, 3 bytes flags
	"minf": 0,
	"moov": 0,
	"stbl": 0,
	"stsd": 8, //1 byte version, 3 bytes flags, 4 bytes number of entries
	"trak": 0,
	"udta": 0,
}

func (m *MP4Metadata) readAtoms(r io.ReadSeeker) error {
	//Get first header, verify it is ftyp atom
	name, size, err := readAtomHeader(r)
	if err != nil {
		return err
	}
	if name != "ftyp" {
		return errors.New("Did not find ftyp atom when reading MP4. Unknown format.")
	}
	//Get major brand (i.e. file type)
	majorBrand, err := readString(r, 4)
	if err != nil {
		return err
	}
	if majorBrand == "M4A " {
		m.fileType = M4A
	} else if majorBrand == "M4B " {
		m.fileType = M4B
	} else if majorBrand == "M4P " {
		m.fileType = M4P
	}
	_, err = r.Seek(int64(size-12), io.SeekCurrent)
	if err != nil {
		return err
	}

	for {
		//All Atoms start with size and name
		name, size, err = readAtomHeader(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		//Check to see if this is a parent atom where we want to read the children
		//and if so how many bytes are between the name and the first child
		skipBytes, ok := parentAtomNamesSkipBytes[name]
		//If so skip only those bytes (not the entire size) and restart the loop
		if ok {
			if skipBytes > 0 {
				_, err = r.Seek(skipBytes, io.SeekCurrent)
			}
			if err != nil {
				return err
			}
			continue
		} else if name == "mvhd" {
			//The mvhd atom has information that will allow us to detect the duration
			//of the file in seconds
			//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap2/qtff2.html#//apple_ref/doc/uid/TP40000939-CH204-56313
			mvhd := MovieHeaderAtom{}
			b, err := readBytes(r, uint(size-8))
			if err != nil {
				return err
			} else if len(b) < 20 {
				return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 20, len(b))
			}
			vers := b[0]
			if vers != 0 {
				//Unknown version, don't process
				continue
			}
			//Next 11 bytes are Flags (3B), Creation Time (4B), and Modification Time (4B)
			mvhd.TimeScale = getInt(b[12:16])
			mvhd.Duration = getInt(b[16:20])
			m.movieHeaderAtom = &mvhd
			continue
		} else if name == "mp4a" {
			//mp4a atom is a sample description stored as a child of a sample
			//description atom (stsd) and it contains channel and sample rate info and
			//also likely has a child esds atom that we need info from as well.
			//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap2/qtff2.html#//apple_ref/doc/uid/TP40000939-CH204-61112
			//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-75770
			if size < 28 {
				return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 28, size)
			}
			mp4a := MP4AAtom{}
			b, err := readBytes(r, 28)
			if err != nil {
				return err
			}
			vers := getInt(b[8:10])
			if vers < 2 {
				mp4a.Channels = getInt(b[16:18])
				mp4a.SampleRate = get16Dot16FixedPointAsFloat(b[24:28])
				//If version is 1 then we have 16 bytes of additional data to ignore
				if vers == 1 {
					if size < 44 {
						return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 44, size)
					}
					//Skip 16 bytes to get the next atom or a child atom
					_, err = r.Seek(16, io.SeekCurrent)
					if err != nil {
						return err
					}
				}
				m.mp4aAtom = &mp4a
			} else if vers == 2 {
				//If version is 2 then the layout of the sample description is different
				//The data we need stil needs to be read
				if size < 64 {
					return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 64, size)
				}
				b2, err := readBytes(r, 36) //read 36 additional bytes
				if err != nil {
					return err
				}
				mp4a.Channels = getInt(b2[12:16])
				mp4a.SampleRate = getFloat64(b2[4:12])
				m.mp4aAtom = &mp4a
			} else {
				//Unknown version, skip to next atom ignoring any children
				_, err = r.Seek(int64(size-28), io.SeekCurrent)
				if err != nil {
					return err
				}
			}
			continue
		} else if name == "esds" {
			//Elementary stream descriptor (likely a child of mp4a atom)
			//Holds bitrate information for the file
			//See generic info at https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-124774
			//For data layout, see section 8.3 in https://www.itscj.ipsj.or.jp/sc29/open/29view/29n2601t.pdf
			//For info on ES descriptor type tags find 'esds' in http://xhelmboyx.tripod.com/formats/mp4-layout.txt
			esds := ESDSAtom{}
			b, err := readBytes(r, uint(size-8))
			if err != nil {
				return err
			} else if len(b) < 8 {
				return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 8, len(b))
			}
			extendedTagSize := 0
			//ES Descriptor Tag of 0x03 (ObjectDescriptor) expected here
			if b[4] != 3 {
				//unknown format, skip
				continue
			}
			//Check if has long version of descriptor tag (look for 0x80 0x80 0x80)
			//If so our offsets will need to be extended by the number of tags
			//encountered before the data, i.e. each tag will be 4 bytes instead of 1
			if getInt(b[5:8]) == 8421504 {
				extendedTagSize = 3
			}
			if len(b) < 24+(2*extendedTagSize) {
				return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 24+(2*extendedTagSize), len(b))
			}
			//ES Descriptor Tag of 0x04 expected here
			if b[9+extendedTagSize] != 4 {
				//unknown format, skip
				continue
			}
			esds.ObjectProfileIndication = b[11+(2*extendedTagSize)]
			esds.StreamType = b[12+(2*extendedTagSize)] >> 2
			esds.MaxBitrate = getInt(b[16+(2*extendedTagSize) : 20+(2*extendedTagSize)])
			esds.AvgBitrate = getInt(b[20+(2*extendedTagSize) : 24+(2*extendedTagSize)])
			m.esdsAtom = &esds
			continue
		} else {
			//Check and see if this is a meta item that we need to process
			ok = containsString(atomMetaItemNames, name)
			var data []string
			if name == "----" {
				name, data, err = readCustomAtom(r, size)
				if err != nil {
					return err
				}

				if name != "----" {
					ok = true
					size = 0 // already read data
				}
			}

			//Not a meta item, skip and move to the next atom
			if !ok {
				_, err := r.Seek(int64(size-8), io.SeekCurrent)
				if err != nil {
					return err
				}
				continue
			}

			err = m.readMetaItemAtomData(r, name, size-8, data)
			if err != nil {
				return err
			}
		}
	}
}

var atomMetaValTypes = map[int]string{
	0:  "implicit", // automatic based on atom name
	1:  "text",
	13: "jpeg",
	14: "png",
	21: "uint8",
}

// Detect PNG image if "implicit" class is used
var pngHeader = []byte{137, 80, 78, 71, 13, 10, 26, 10}

func (m MP4Metadata) readMetaItemAtomData(r io.ReadSeeker, name string, size uint32, processedData []string) error {
	var b []byte
	var err error
	var contentType string
	if len(processedData) > 0 {
		b = []byte(strings.Join(processedData, ";")) // add delimiter if multiple data fields
		contentType = "text"
	} else {
		// read the data
		b, err = readBytes(r, uint(size))
		if err != nil {
			return err
		}
		if len(b) < 8 {
			return fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 8, len(b))
		}

		// "data" + size (4 bytes each)
		b = b[8:]

		if len(b) < 3 {
			return fmt.Errorf("invalid encoding: expected at least %d bytes, for class, got %d", 3, len(b))
		}
		class := getInt(b[1:4])
		var ok bool
		contentType, ok = atomMetaValTypes[class]
		if !ok {
			return fmt.Errorf("invalid content type: %v (%x) (%x)", class, b[1:4], b)
		}

		// 4: atom version (1 byte) + atom flags (3 bytes)
		// 4: NULL (usually locale indicator)
		if len(b) < 8 {
			return fmt.Errorf("invalid encoding: expected at least %d bytes, for atom version and flags, got %d", 8, len(b))
		}
		b = b[8:]
	}

	if name == "trkn" || name == "disk" {
		if len(b) < 6 {
			return fmt.Errorf("invalid encoding: expected at least %d bytes, for track and disk numbers, got %d", 6, len(b))
		}

		m.data[name] = int(b[3])
		m.data[name+"_count"] = int(b[5])
		return nil
	}

	if contentType == "implicit" {
		if name == "covr" {
			if bytes.HasPrefix(b, pngHeader) {
				contentType = "png"
			}
			// TODO(dhowden): Detect JPEG formats too (harder).
		}
	}

	var data interface{}
	switch contentType {
	case "implicit":
		if ok := containsString(atomMetaItemNames, name); ok {
			return fmt.Errorf("unhandled implicit content type for required atom: %q", name)
		}
		return nil

	case "text":
		data = string(b)

	case "uint8":
		if len(b) < 1 {
			return fmt.Errorf("invalid encoding: expected at least %d bytes, for integer tag data, got %d", 1, len(b))
		}
		data = getInt(b[:1])

	case "jpeg", "png":
		data = &Picture{
			Ext:      contentType,
			MIMEType: "image/" + contentType,
			Data:     b,
		}
	}
	m.data[name] = data

	return nil
}

func readAtomHeader(r io.ReadSeeker) (name string, size uint32, err error) {
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return
	}
	name, err = readString(r, 4)
	return
}

// Generic atom.
// Should have 3 sub atoms : mean, name and data.
// We check that mean is "com.apple.iTunes" and we use the subname as
// the name, and move to the data atom.
// Data atom could have multiple data values, each with a header.
// If anything goes wrong, we jump at the end of the "----" atom.
func readCustomAtom(r io.ReadSeeker, size uint32) (_ string, data []string, _ error) {
	subNames := make(map[string]string)

	for size > 8 {
		subName, subSize, err := readAtomHeader(r)
		if err != nil {
			return "", nil, err
		}

		// Remove the size of the atom from the size counter
		if size >= subSize {
			size -= subSize
		} else {
			return "", nil, errors.New("--- invalid size")
		}

		b, err := readBytes(r, uint(subSize-8))
		if err != nil {
			return "", nil, err
		}

		if len(b) < 4 {
			return "", nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 4, len(b))
		}
		switch subName {
		case "mean", "name":
			subNames[subName] = string(b[4:])
		case "data":
			data = append(data, string(b[4:]))
		}
	}

	// there should remain only the header size
	if size != 8 {
		err := errors.New("---- atom out of bounds")
		return "", nil, err
	}

	if subNames["mean"] != "com.apple.iTunes" || subNames["name"] == "" || len(data) == 0 {
		return "----", nil, nil
	}
	return subNames["name"], data, nil
}

func (m MP4Metadata) Album() string {
	t, ok := m.data["\xa9alb"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) AlbumArtist() string {
	t, ok := m.data["aART"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) Artist() string {
	t, ok := m.data["\xa9art"]
	if !ok {
		t, ok = m.data["\xa9ART"]
		if !ok {
			return ""
		}
	}
	return t.(string)
}

func (m MP4Metadata) AverageBitrate() int {
	if m.esdsAtom != nil {
		if m.esdsAtom.AvgBitrate != 0 {
			return m.esdsAtom.AvgBitrate
		}
		return m.esdsAtom.MaxBitrate
	}
	return 0
}

func (m MP4Metadata) Comment() string {
	t, ok := m.data["\xa9cmt"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) Composer() string {
	t, ok := m.data["\xa9wrt"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) Disc() (int, int) {
	var x, y = 0, 0
	if xi, ok := m.data["disk"]; ok {
		x = xi.(int)
	}
	if yi, ok := m.data["disk_count"]; ok {
		y = yi.(int)
	}
	return x, y
}

func (m MP4Metadata) Duration() time.Duration {
	if m.movieHeaderAtom == nil {
		return time.Duration(0)
	}
	//Calculate true duration by dividing duration (total samples) by time scale (sample rate)
	seconds := float64(m.movieHeaderAtom.Duration) / float64(m.movieHeaderAtom.TimeScale)
	return time.Duration(seconds * float64(time.Second))
}

//Returns information extracted from the elementary stream descriptor atom
//('esds') found in the file.
func (m MP4Metadata) ESDSAtom() *ESDSAtom {
	return m.esdsAtom
}

func (m MP4Metadata) FileType() FileType {
	return m.fileType
}

func (m MP4Metadata) Format() Format {
	if m.data != nil {
		return MP4
	}
	return UnknownFormat
}

func (m MP4Metadata) Genre() string {
	t, ok := m.data["\xa9gen"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) Lyrics() string {
	t, ok := m.data["\xa9lyr"]
	if !ok {
		return ""
	}
	return t.(string)
}

//Returns information extracted from the movie header atom ('mvhd') found in the
//file.
func (m MP4Metadata) MovieHeaderAtom() *MovieHeaderAtom {
	return m.movieHeaderAtom
}

//Returns information extracted from the MP4A sound sample description atom
//('mp4a') found in the file.
func (m MP4Metadata) MP4AAtom() *MP4AAtom {
	return m.mp4aAtom
}

func (m MP4Metadata) Picture() *Picture {
	v, ok := m.data["covr"]
	if !ok {
		return nil
	}
	p, _ := v.(*Picture)
	return p
}

func (m MP4Metadata) Raw() map[string]interface{} {
	return m.data
}

func (m MP4Metadata) Title() string {
	t, ok := m.data["\xa9nam"]
	if !ok {
		return ""
	}
	return t.(string)
}

func (m MP4Metadata) Track() (int, int) {
	var x, y = 0, 0
	if xi, ok := m.data["trkn"]; ok {
		x = xi.(int)
	}
	if yi, ok := m.data["trkn_count"]; ok {
		y = yi.(int)
	}
	return x, y
}

func (m MP4Metadata) Year() int {
	t, ok := m.data["\xa9day"]
	if !ok {
		return 0
	}
	date := t.(string)
	if len(date) >= 4 {
		year, _ := strconv.Atoi(date[:4])
		return year
	}
	return 0
}*/
