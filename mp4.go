package yurit

import (
	"encoding/binary"
	"io"
	"time"
)

type Mp4Atom struct {
	Name     string
	Size     uint32
	Data     []byte
	Children []Mp4Atom
}

var parentAtomsList = map[string]int64{
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

var childrenAreParentsAtomsList = []string{
	"ilst",
}

var dataAtomsList = []string{
	"data",
	"esds",
	"ftyp",
	"mean",
	"mvhd",
	"mp4a",
	"name",
}

// MP4Metadata is the implementation of Metadata for MP4 tag (atom) data.
type MP4Metadata struct {
	ftyp     mp4ftyp
	esds     mp4esds
	metadata mp4metadata
	mp4a     mp4mp4a
	mvhd     mp4mvhd
}

// ReadMP4 reads MP4 metadata atoms from the io.ReadSeeker into a Metadata, returning
// non-nil error if there was a problem.
func ReadMP4(r io.ReadSeeker) (*MP4Metadata, error) {
	a, err := readMp4AtomsFunc(r, true, 0, false)
	if err != nil {
		return nil, err
	}
	m := MP4Metadata{}
	ftypAtom := findAtom(a, "ftyp")
	if ftypAtom != nil {
		m.ftyp, err = processFTYPAtom(*ftypAtom)
		if err != nil {
			return nil, err
		}
	}
	mvhdAtom := findAtom(a, "mvhd")
	if mvhdAtom != nil {
		m.mvhd, err = processMVHDAtom(*mvhdAtom)
		if err != nil {
			return nil, err
		}
	}
	mp4aAtom := findAtom(a, "mp4a")
	if mp4aAtom != nil {
		m.mp4a, m.esds, err = processMP4AAtom(*mp4aAtom)
		if err != nil {
			return nil, err
		}
	}
	ilstAtom := findAtom(a, "ilst")
	if ilstAtom != nil {
		m.metadata, err = processILSTAtom(*ilstAtom)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

//readMp4AtomsFunc reads an MP4 file and returns a slice of Mp4Atom structs representing
//the (partial) layout of atoms within the MP4. Each Mp4Atom parent may have its children
//listed under its Children variable, but only those parent atoms listed in the
//parentAtomsList are searched. If an atom is listed in the dataAtomsList then
//its data is added to the Mp4Atom's Data variable.
func readMp4AtomsFunc(r io.ReadSeeker, readAll bool, readSize int, forceParent bool) ([]Mp4Atom, error) {
	var a []Mp4Atom
	//If set to read all or if we haven't read the entire specified size, keep
	//reading atoms.
	for readAll || readSize > 0 {
		//Get atom header
		name, size, err := readAtomHeader(r)
		if err != nil {
			if err == io.EOF {
				return a, nil
			}
			return nil, err
		}
		//Determine if this is a parent atom, and if so, how many bytes need to be
		//skipped to get to the first child atom header. Usually this is 0, but not
		//always.
		skipBytes, ok := parentAtomsList[name]
		if ok || forceParent {
			//Specified parent atom...
			//If we have bytes to skip, skip them
			if skipBytes > 0 {
				_, err = r.Seek(skipBytes, io.SeekCurrent)
			}
			if err != nil {
				return nil, err
			}
			childrenAreParents := containsString(childrenAreParentsAtomsList, name)
			//Recursively call read func to read children, but stop once we get to the
			//end of the parent
			children, err := readMp4AtomsFunc(r, false, (int(size) - (8 + int(skipBytes))), childrenAreParents)
			if err != nil {
				return nil, err
			}
			//Add this atom along with its children to the slice
			a = append(a, Mp4Atom{Name: name, Size: size, Children: children})
		} else {
			//Either a data atom or a parent atom where we don't care about the children...
			var data []byte
			//If this is in our data atoms list or we've been commanded to read all data,
			//grab its data, else just seek past it to the start of the next atom
			if containsString(dataAtomsList, name) {
				data, err = readBytes(r, uint(size-8))
			} else {
				_, err = r.Seek(int64(size-8), io.SeekCurrent)
			}
			if err != nil {
				return nil, err
			}
			//Add this atom to the slice
			a = append(a, Mp4Atom{Name: name, Size: size, Data: data})
		}
		//Reduce readSize by the size of this atom which should now be read
		readSize -= int(size)
	}
	return a, nil
}

func readAtomHeader(r io.ReadSeeker) (name string, size uint32, err error) {
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return
	}
	name, err = readString(r, 4)
	return
}

func findAtom(atoms []Mp4Atom, name string) *Mp4Atom {
	for _, atom := range atoms {
		if atom.Name == name {
			return &atom
		}
		if len(atom.Children) > 0 {
			child := findAtom(atom.Children, name)
			if child != nil {
				return child
			}
		}
	}
	return nil
}

func (m MP4Metadata) Album() string {
	return m.metadata.Album()
}

func (m MP4Metadata) AlbumArtist() string {
	return m.metadata.AlbumArtist()
}

func (m MP4Metadata) Artist() string {
	return m.metadata.Artist()
}

func (m MP4Metadata) AverageBitrate() int {
	return m.esds.AverageBitrate()
}

func (m MP4Metadata) Comment() string {
	return m.metadata.Comment()
}

func (m MP4Metadata) Composer() string {
	return m.metadata.Composer()
}

func (m MP4Metadata) Disc() (int, int) {
	return m.metadata.Disc()
}

func (m MP4Metadata) Duration() time.Duration {
	return m.mvhd.Duration()
}

//Returns information extracted from the elementary stream descriptor atom
//('esds') found in the file.
func (m MP4Metadata) ESDS() map[string]interface{} {
	return m.esds
}

func (m MP4Metadata) FileType() FileType {
	return m.ftyp.FileType()
}

func (m MP4Metadata) Format() Format {
	if m.metadata != nil {
		return MP4
	}
	return UnknownFormat
}

func (m MP4Metadata) FTYP() map[string]interface{} {
	return m.ftyp
}

func (m MP4Metadata) Genre() string {
	return m.metadata.Genre()
}

func (m MP4Metadata) Lyrics() string {
	return m.metadata.Lyrics()
}

//Returns information extracted from the MP4A sound sample description atom
//('mp4a') found in the file.
func (m MP4Metadata) MP4A() map[string]interface{} {
	return m.mp4a
}

//Returns information extracted from the movie header atom ('mvhd') found in the
//file.
func (m MP4Metadata) MVHD() map[string]interface{} {
	return m.mvhd
}

func (m MP4Metadata) Picture() *Picture {
	return m.metadata.Picture()
}

func (m MP4Metadata) Raw() map[string]interface{} {
	return m.metadata
}

func (m MP4Metadata) Title() string {
	return m.metadata.Title()
}

func (m MP4Metadata) Track() (int, int) {
	return m.metadata.Track()
}

func (m MP4Metadata) Year() int {
	return m.metadata.Year()
}
