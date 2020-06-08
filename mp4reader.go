package yurit

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
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
	ftypMap         FTYPMap
	esdsMap         ESDSMap
	metadataItemMap Mp4MetadataItemMap
	mp4aMap         MP4AMap
	mvhdMap         MVHDMap
}

type ESDSMap map[string]interface{}

type FTYPMap map[string]interface{}

type Mp4MetadataItemMap map[string]interface{}

type MP4AMap map[string]interface{}

type MVHDMap map[string]interface{}

// ReadMP4 reads MP4 metadata atoms from the io.ReadSeeker into a Metadata, returning
// non-nil error if there was a problem.
func ReadMP4(r io.ReadSeeker) (*MP4Metadata, error) {
	a, err := ReadMp4Atoms(r)
	if err != nil {
		return nil, err
	}
	m := MP4Metadata{}
	ftypAtom := findAtom(a, "ftyp")
	if ftypAtom != nil {
		ftyp, err := ProcessFTYPAtom(*ftypAtom)
		if err != nil {
			return nil, err
		}
		m.ftypMap = ftyp
	}
	mvhdAtom := findAtom(a, "mvhd")
	if mvhdAtom != nil {
		mvhd, err := ProcessMVHDAtom(mvhdAtom.Data)
		if err != nil {
			return nil, err
		}
		m.mvhdMap = mvhd
	}
	mp4aAtom := findAtom(a, "mp4a")
	if mp4aAtom != nil {
		mp4a, esds, err := ProcessMP4AAtom(mp4aAtom.Data)
		if err != nil {
			return nil, err
		}
		m.mp4aMap = mp4a
		m.esdsMap = esds
	}
	ilstAtom := findAtom(a, "ilst")
	if ilstAtom != nil {
		metaItemMap, err := ProcessILSTAtom(*ilstAtom)
		if err != nil {
			return nil, err
		}
		m.metadataItemMap = metaItemMap
	}
	return &m, nil
}

func ReadMp4Atoms(r io.ReadSeeker) ([]Mp4Atom, error) {
	a, err := readMp4AtomsFunc(r, true, 0, false)
	if err != nil {
		return nil, err
	}
	return a, nil
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

func ProcessMP4AAtom(b []byte) (MP4AMap, ESDSMap, error) {
	//mp4a atom is a sample description stored as a child of a sample
	//description atom (stsd) and it contains channel and sample rate info and
	//also likely has a child esds atom that we need info from as well.
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap2/qtff2.html#//apple_ref/doc/uid/TP40000939-CH204-61112
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-75770
	if len(b) < 28 {
		return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 28, len(b))
	}
	mp4a := MP4AMap{}
	childOffset := 28
	//Skip 6 reserved bytes and the data reference index key (2 bytes)
	version := getUint16(b[8:10])
	mp4a[VersionKey] = version
	mp4a[RevisionKey] = getUint16(b[10:12])
	mp4a[VendorKey] = getUint32(b[12:16])
	if version < 2 {
		mp4a[ChannelsKey] = getUint16(b[16:18])
		mp4a[SampleSizeKey] = getUint16(b[18:20])
		mp4a[CompressionKey] = getInt16(b[20:22])
		mp4a[PacketSizeKey] = getUint16(b[22:24])
		mp4a[SampleRateKey] = get16Dot16FixedPointAsFloat(b[24:28])
		if version == 1 {
			if len(b) < 44 {
				return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 44, len(b))
			}
			childOffset = 44
			mp4a[SamplesPerPacketKey] = getUint32(b[28:32])
			mp4a[BytesPerPacketKey] = getUint32(b[32:36])
			mp4a[BytesPerFrameKey] = getUint32(b[36:40])
			mp4a[BytesPerSampleKey] = getUint32(b[40:44])
		}
	} else if version == 2 {
		//If version is 2 then the layout of the sample description is different
		if len(b) < 64 {
			return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 64, len(b))
		}
		childOffset = 64
		mp4a[Always3Key] = getInt16(b[16:18])
		mp4a[Always16Key] = getInt16(b[18:20])
		mp4a[AlwaysMinus2Key] = getInt16(b[20:22])
		mp4a[Always0Key] = getInt16(b[22:24])
		mp4a[Always65536Key] = getInt32(b[24:28])
		mp4a[SizeOfStructOnlyKey] = getInt32(b[28:32])
		mp4a[SampleRateKey] = getFloat64(b[32:40])
		mp4a[ChannelsKey] = getInt32(b[40:44])
		mp4a[Always7F000000Key] = getInt32(b[44:48])
		mp4a[ConstBitsPerChannelKey] = getInt32(b[48:52])
		mp4a[LPCMFlagsKey] = getInt32(b[52:56])
		mp4a[ConstBytesPerPacket] = getUint32(b[56:60])
		mp4a[ConstFramesPerPacket] = getUint32(b[60:64])
	} else {
		return nil, nil, errors.New("Unknown version.")
	}
	childReader := bytes.NewReader(b[childOffset:])
	children, err := readMp4AtomsFunc(childReader, true, 0, false)
	if err != nil {
		return mp4a, nil, err
	}
	esdsAtom := findAtom(children, "esds")
	if esdsAtom == nil {
		return mp4a, nil, nil
	}
	esds, err := ProcessESDSAtom(esdsAtom.Data)
	if err != nil {
		return mp4a, nil, err
	}
	return mp4a, esds, nil
}

func ProcessESDSAtom(b []byte) (ESDSMap, error) {
	if len(b) < 29 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29, len(b))
	}
	esds := ESDSMap{}
	s := 0 //extended tag size
	esds[VersionKey] = getUint32(b[0:4])
	esds["esDescriptorTypeTag"] = b[4]
	if b[5] == 0x80 && b[6] == 0x80 && b[7] == 0x80 {
		s = 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds["esDescriptorLength"] = b[5+s]
	esds["esID"] = getUint16(b[6+s : 8+s])
	esds["streamPriority"] = b[8+s]
	esds["decoderConfigDescriptorTypeTag"] = b[9+s]
	if b[10+s] == 0x80 && b[11+s] == 0x80 && b[12+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds["decoderConfigDescriptorLength"] = b[10+s]
	esds["objectTypeID"] = b[11+s]
	esds["streamType"] = b[12+s] >> 2
	esds["upstream"] = getBit(b[12+s], 1)
	esds["reserved"] = getBit(b[12+s], 0)
	esds["bufferSize"] = getUint24(b[13+s : 16+s])
	esds[MaximumBitrateKey] = getUint32(b[16+s : 20+s])
	esds[AverageBitrateKey] = getUint32(b[20+s : 24+s])
	esds["decoderSpecificDescriptorTypeTag"] = b[24+s]
	if b[25+s] == 0x80 && b[26+s] == 0x80 && b[27+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds["decoderSpecificDescriptorLength"] = b[25+s]
	esds["decoderSpecificInfo"] = b[26+s : 26+s+int(esds["decoderSpecificDescriptorLength"].(uint8))]
	s += int(esds["decoderSpecificDescriptorLength"].(uint8))
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds["slConfigDescriptorTypeTag"] = b[26+s]
	if b[27+s] == 0x80 && b[28+s] == 0x80 && b[29+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds["slConfigDescriptorLength"] = b[27+s]
	esds["slValue"] = b[28+s]
	return esds, nil
}

func ProcessMVHDAtom(b []byte) (MVHDMap, error) {
	if len(b) < 100 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 100, len(b))
	}
	mvhd := MVHDMap{}
	mvhd[VersionKey] = b[0]
	mvhd[FlagsKey] = b[1:4]
	mvhd["creationTime"] = getUint32(b[4:8])
	mvhd["modificationTime"] = getUint32(b[8:12])
	mvhd[TimeScaleKey] = getUint32(b[12:16])
	mvhd[DurationKey] = getUint32(b[16:20])
	mvhd["preferredRate"] = get16Dot16FixedPointAsFloat(b[20:24])
	mvhd["preferredVolume"] = get8Dot8FixedPointAsFloat(b[24:26])
	//Skip 10 bytes reserved
	mvhd["matrixA"] = get16Dot16FixedPointAsFloat(b[36:40])
	mvhd["matrixB"] = get16Dot16FixedPointAsFloat(b[40:44])
	mvhd["matrixU"] = get16Dot16FixedPointAsFloat(b[44:48])
	mvhd["matrixC"] = get16Dot16FixedPointAsFloat(b[48:52])
	mvhd["matrixD"] = get16Dot16FixedPointAsFloat(b[52:56])
	mvhd["matrixV"] = get16Dot16FixedPointAsFloat(b[56:60])
	mvhd["matrixX"] = get16Dot16FixedPointAsFloat(b[60:64])
	mvhd["matrixY"] = get16Dot16FixedPointAsFloat(b[64:68])
	mvhd["matrixW"] = get16Dot16FixedPointAsFloat(b[68:72])
	mvhd["previewTime"] = getUint32(b[72:76])
	mvhd["previewDuration"] = getUint32(b[76:80])
	mvhd["posterTime"] = getUint32(b[80:84])
	mvhd["selectionTime"] = getUint32(b[84:88])
	mvhd["selectionDuration"] = getUint32(b[88:92])
	mvhd["currentTime"] = getUint32(b[92:96])
	mvhd["nextTrackID"] = getInt32(b[96:100])
	return mvhd, nil
}

func ProcessILSTAtom(ilst Mp4Atom) (Mp4MetadataItemMap, error) {
	m := Mp4MetadataItemMap{}
	for _, metadataItem := range ilst.Children {
		name := metadataItem.Name
		if name == "----" {
			mean := ""
			meanAtom := findAtom(metadataItem.Children, "mean")
			if meanAtom != nil && len(meanAtom.Data) > 4 {
				//Skip version/flag bytes, and just get mean data as string
				mean = string(meanAtom.Data[4:])
			}
			nameAtom := findAtom(metadataItem.Children, "name")
			if nameAtom != nil && len(nameAtom.Data) > 4 {
				//Skip version/flag bytes, and just get name data as string
				name = string(nameAtom.Data[4:])
			}
			//Verify mean is expected value and name has changed
			if mean != "com.apple.iTunes" || name == "----" {
				continue
			}
		}
		//Find first data atom
		dataAtom := findAtom(metadataItem.Children, "data")
		if len(dataAtom.Data) < 8 {
			return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for atom version and flags, got %d", 8, len(dataAtom.Data))
		}
		dataType := getUint24(dataAtom.Data[1:4])
		dataPortion := dataAtom.Data[8:]
		if name == "trkn" || name == "disk" {
			if len(dataPortion) < 6 {
				return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for track and disk numbers, got %d", 6, len(dataPortion))
			}
			m[name] = int64(dataPortion[3])
			m[name+"_count"] = int64(dataPortion[5])
		} else if name == "covr" {
			contentType := ""
			if dataType == 13 {
				contentType = "jpeg"
			} else if dataType == 14 || bytes.HasPrefix(dataPortion, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
				contentType = "png"
			}
			if contentType == "" {
				continue
			}
			picture := &Picture{
				Ext:      contentType,
				MIMEType: "image/" + contentType,
				Data:     dataPortion,
			}
			m[name] = picture
		} else if dataType == 0 {
			m[name] = dataPortion
		} else if dataType == 1 {
			m[name] = string(dataPortion)
		} else if dataType == 21 {
			if len(dataPortion) < 1 {
				return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for integer tag data, got %d", 1, len(dataPortion))
			}
			m[name] = int64(dataPortion[0])
		}
	}
	return m, nil
}

func ProcessFTYPAtom(ftyp Mp4Atom) (FTYPMap, error) {
	if len(ftyp.Data) < 12 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 12, len(ftyp.Data))
	}
	m := FTYPMap{}
	m[MajorBrandKey] = string(ftyp.Data[:4])
	m["minorVersion"] = ftyp.Data[4:8]
	cb := []string{string(ftyp.Data[8:12])}
	for i := 12; i+4 < len(ftyp.Data); i += 4 {
		cb = append(cb, string(ftyp.Data[i:i+4]))
	}
	m["compatibleBrands"] = cb
	return m, nil
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
	t, _ := m.metadataItemMap["\xa9alb"].(string)
	return t
}

func (m MP4Metadata) AlbumArtist() string {
	t, _ := m.metadataItemMap["aART"].(string)
	return t
}

func (m MP4Metadata) Artist() string {
	t, ok := m.metadataItemMap["\xa9art"].(string)
	if !ok {
		t, _ = m.metadataItemMap["\xa9ART"].(string)
	}
	return t
}

func (m MP4Metadata) AverageBitrate() int64 {
	if m.esdsMap != nil {
		b, ok := m.esdsMap[AverageBitrateKey].(int64)
		if !ok || b == 0 {
			b, _ = m.esdsMap[MaximumBitrateKey].(int64)
		}
		return b
	}
	return 0
}

func (m MP4Metadata) Comment() string {
	t, _ := m.metadataItemMap["\xa9cmt"].(string)
	return t
}

func (m MP4Metadata) Composer() string {
	t, _ := m.metadataItemMap["\xa9wrt"].(string)
	return t
}

func (m MP4Metadata) Disc() (int64, int64) {
	x, _ := m.metadataItemMap["disk"].(int64)
	y, _ := m.metadataItemMap["disk_count"].(int64)
	return x, y
}

func (m MP4Metadata) Duration() time.Duration {
	//Chck that the mvhd atom map exists and we have the required values
	if m.mvhdMap != nil {
		d, okd := m.mvhdMap[DurationKey].(int64)
		t, okt := m.mvhdMap[TimeScaleKey].(int64)
		if okd && okt && t != 0 {
			//Calculate true duration by dividing duration (total samples) by time scale (sample rate)
			seconds := float64(d) / float64(t)
			return time.Duration(seconds * float64(time.Second))
		}
	}
	//Missing data, cannot calculate, return 0 as a duration
	return time.Duration(0)
}

//Returns information extracted from the elementary stream descriptor atom
//('esds') found in the file.
func (m MP4Metadata) ESDS() ESDSMap {
	return m.esdsMap
}

func (m MP4Metadata) FileType() FileType {
	f := UnknownFileType
	if m.ftypMap != nil {
		mb, ok := m.ftypMap[MajorBrandKey].(string)
		if ok {
			if mb == "M4A " {
				f = M4A
			} else if mb == "M4B " {
				f = M4B
			} else if mb == "M4P " {
				f = M4P
			}
		}
	}
	return f
}

func (m MP4Metadata) Format() Format {
	if m.metadataItemMap != nil {
		return MP4
	}
	return UnknownFormat
}

func (m MP4Metadata) FTYP() FTYPMap {
	return m.ftypMap
}

func (m MP4Metadata) Genre() string {
	t, _ := m.metadataItemMap["\xa9gen"].(string)
	return t
}

func (m MP4Metadata) Lyrics() string {
	t, _ := m.metadataItemMap["\xa9lyr"].(string)
	return t
}

//Returns information extracted from the MP4A sound sample description atom
//('mp4a') found in the file.
func (m MP4Metadata) MP4A() MP4AMap {
	return m.mp4aMap
}

//Returns information extracted from the movie header atom ('mvhd') found in the
//file.
func (m MP4Metadata) MVHD() MVHDMap {
	return m.mvhdMap
}

func (m MP4Metadata) Picture() *Picture {
	p, _ := m.metadataItemMap["covr"].(*Picture)
	return p
}

func (m MP4Metadata) Raw() map[string]interface{} {
	return m.metadataItemMap
}

func (m MP4Metadata) Title() string {
	t, _ := m.metadataItemMap["\xa9nam"].(string)
	return t
}

func (m MP4Metadata) Track() (int64, int64) {
	x, _ := m.metadataItemMap["trkn"].(int64)
	y, _ := m.metadataItemMap["trkn_count"].(int64)
	return x, y
}

func (m MP4Metadata) Year() int64 {
	var year int64 = 0
	t, ok := m.metadataItemMap["\xa9day"].(string)
	if ok && len(t) >= 4 {
		year, _ = strconv.ParseInt(t[:4], 10, 64)
	}
	return year
}
