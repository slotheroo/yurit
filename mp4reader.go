package yurit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

var childDataAtomsList = []string{
	"ilst",
}

var dataAtomsList = []string{
	"esds",
	"mvhd",
	"mp4a",
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
func readMp4AtomsFunc(r io.ReadSeeker, readAll bool, readSize int, allData bool) ([]Mp4Atom, error) {
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
		if ok {
			//Specified parent atom...
			//If we have bytes to skip, skip them
			if skipBytes > 0 {
				_, err = r.Seek(skipBytes, io.SeekCurrent)
			}
			if err != nil {
				return nil, err
			}
			needChildData := containsString(childDataAtomsList, name)
			//Recursively call read func to read children, but stop once we get to the
			//end of the parent
			children, err := readMp4AtomsFunc(r, false, (int(size) - (8 + int(skipBytes))), needChildData)
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
			if allData || containsString(dataAtomsList, name) {
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

func ProcessMP4AAtom(b []byte) (*MP4ASampleDescription, *ESDSExtension, error) {
	//mp4a atom is a sample description stored as a child of a sample
	//description atom (stsd) and it contains channel and sample rate info and
	//also likely has a child esds atom that we need info from as well.
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap2/qtff2.html#//apple_ref/doc/uid/TP40000939-CH204-61112
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-75770
	if len(b) < 28 {
		return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 28, len(b))
	}
	mp4a := MP4ASampleDescription{raw: map[string]interface{}{}}
	childOffset := 28
	//Skip 6 reserved bytes and the data reference index key (2 bytes)
	version := getInt16(b[8:10])
	mp4a.raw[VersionKey] = version
	mp4a.raw[RevisionKey] = getInt16(b[10:12])
	mp4a.raw[VendorKey] = getInt32(b[12:16])
	if version < 2 {
		mp4a.raw[ChannelsKey] = getInt16(b[16:18])
		mp4a.raw[SampleSizeKey] = getInt16(b[18:20])
		mp4a.raw[CompressionKey] = getInt16(b[20:22])
		mp4a.raw[PacketSizeKey] = getInt16(b[22:24])
		mp4a.raw[SampleRateKey] = get16Dot16FixedPointAsFloat(b[24:28])
		if version == 1 {
			if len(b) < 44 {
				return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 44, len(b))
			}
			childOffset = 44
			mp4a.raw[SamplesPerPacketKey] = getInt32(b[28:32])
			mp4a.raw[BytesPerPacketKey] = getInt32(b[32:36])
			mp4a.raw[BytesPerFrameKey] = getInt32(b[36:40])
			mp4a.raw[BytesPerSampleKey] = getInt32(b[40:44])
		}
	} else if version == 2 {
		//If version is 2 then the layout of the sample description is different
		if len(b) < 64 {
			return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 64, len(b))
		}
		childOffset = 64
		mp4a.raw[Always3Key] = getInt16(b[16:18])
		mp4a.raw[Always16Key] = getInt16(b[18:20])
		mp4a.raw[AlwaysMinus2Key] = getInt16(b[20:22])
		mp4a.raw[Always0Key] = getInt16(b[22:24])
		mp4a.raw[Always65536Key] = getInt32(b[24:28])
		mp4a.raw[SizeOfStructOnlyKey] = getInt32(b[28:32])
		mp4a.raw[SampleRateKey] = getFloat64(b[32:40])
		mp4a.raw[ChannelsKey] = getInt32(b[40:44])
		mp4a.raw[Always7F000000Key] = getInt32(b[44:48])
		mp4a.raw[ConstBitsPerChannelKey] = getInt32(b[48:52])
		mp4a.raw[LPCMFlagsKey] = getInt32(b[52:56])
		mp4a.raw[ConstBytesPerPacket] = getUint32(b[56:60])
		mp4a.raw[ConstFramesPerPacket] = getUint32(b[60:64])
	} else {
		return nil, nil, errors.New("Unknown version.")
	}
	childReader := bytes.NewReader(b[childOffset:])
	children, err := readMp4AtomsFunc(childReader, true, 0, false)
	if err != nil {
		return &mp4a, nil, err
	}
	esdsAtom := findAtom(children, "esds")
	if esdsAtom == nil {
		return &mp4a, nil, nil
	}
	esds, err := ProcessESDSAtom(esdsAtom.Data)
	if err != nil {
		return &mp4a, nil, err
	}
	return &mp4a, esds, nil
}

func ProcessESDSAtom(b []byte) (*ESDSExtension, error) {
	if len(b) < 29 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29, len(b))
	}
	esds := ESDSExtension{raw: map[string]interface{}{}}
	s := 0 //extended tag size
	esds.raw[VersionKey] = getUint32(b[0:4])
	esds.raw["esDescriptorTypeTag"] = int64(b[4])
	if b[5] == 0x80 && b[6] == 0x80 && b[7] == 0x80 {
		s = 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds.raw["esDescriptorLength"] = int64(b[5+s])
	esds.raw["esID"] = getUint16(b[6+s : 8+s])
	esds.raw["streamPriority"] = int64(b[8+s])
	esds.raw["decoderConfigDescriptorTypeTag"] = int64(b[9+s])
	if b[10+s] == 0x80 && b[11+s] == 0x80 && b[12+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds.raw["decoderConfigDescriptorLength"] = int64(b[10+s])
	esds.raw["objectTypeID"] = int64(b[11+s])
	esds.raw["streamType"] = int64(b[12+s] >> 2)
	esds.raw["upstream"] = getBit(b[12+s], 1)
	esds.raw["reserved"] = getBit(b[12+s], 0)
	esds.raw["bufferSize"] = getUint24(b[13+s : 16+s])
	esds.raw[MaximumBitrateKey] = getUint32(b[16+s : 20+s])
	esds.raw[AverageBitrateKey] = getUint32(b[20+s : 24+s])
	esds.raw["decoderSpecificDescriptorTypeTag"] = int64(b[24+s])
	if b[25+s] == 0x80 && b[26+s] == 0x80 && b[27+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds.raw["decoderSpecificDescriptorLength"] = int64(b[25+s])
	esds.raw["decoderSpecificInfo"] = b[26+s : 26+s+int(esds.raw["decoderSpecificDescriptorLength"].(int64))]
	s += int(esds.raw["decoderSpecificDescriptorLength"].(int64))
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds.raw["slConfigDescriptorTypeTag"] = int64(b[26+s])
	if b[27+s] == 0x80 && b[28+s] == 0x80 && b[29+s] == 0x80 {
		s += 3
	}
	if len(b) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(b))
	}
	esds.raw["slConfigDescriptorLength"] = int64(b[27+s])
	esds.raw["slValue"] = int64(b[28+s])
	fmt.Println("end", 28+s+1)
	return &esds, nil
}

/*fmt.Println("esds", string(b[32:36]))
fmt.Println("vers", getInt(b[36:37]))
fmt.Println("flags", getInt(b[37:40]))
fmt.Println("desc type flg", getInt(b[40:41]))
t := 0
if getInt(b[41:44]) == 8421504 {
	t = 3
}
fmt.Println("desc type len", getInt(b[41+t:42+t]))
fmt.Println("ES ID", getInt(b[42+t:44+t]))
fmt.Println("priority", getInt(b[44+t:45+t]))
fmt.Println("dec conf desc type", getInt(b[45+t:46+t]))
fmt.Println("dec conf desc len", getInt(b[46+2*t:47+2*t]))
fmt.Println("obj type", b[47+2*t])
fmt.Println("stream type", b[48+2*t]>>2)
fmt.Println("max bitrate", getInt(b[52+2*t:56+2*t]))
fmt.Println("avg bitrate", getInt(b[56+2*t:60+2*t]))*/

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
