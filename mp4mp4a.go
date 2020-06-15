package yurit

import (
	"bytes"
	"errors"
	"fmt"
)

type mp4mp4a map[string]interface{}

func processMP4AAtom(mp4aAtom Mp4Atom) (mp4mp4a, mp4esds, error) {
	//mp4a atom is a sample description stored as a child of a sample
	//description atom (stsd) and it contains channel and sample rate info and
	//also likely has a child esds atom that we need info from as well.
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap2/qtff2.html#//apple_ref/doc/uid/TP40000939-CH204-61112
	//https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-75770
	if len(mp4aAtom.Data) < 28 {
		return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 28, len(mp4aAtom.Data))
	}
	mp4a := mp4mp4a{}
	childOffset := 28
	//Skip 6 reserved bytes and the data reference index key (2 bytes)
	version := getUint16AsInt(mp4aAtom.Data[8:10])
	mp4a[VersionKey] = version
	mp4a[RevisionKey] = getUint16AsInt(mp4aAtom.Data[10:12])
	mp4a[VendorKey] = getUint32AsInt64(mp4aAtom.Data[12:16])
	if version < 2 {
		mp4a[ChannelsKey] = getUint16AsInt(mp4aAtom.Data[16:18])
		mp4a[SampleSizeKey] = getUint16AsInt(mp4aAtom.Data[18:20])
		mp4a[CompressionKey] = getInt16AsInt(mp4aAtom.Data[20:22])
		mp4a[PacketSizeKey] = getUint16AsInt(mp4aAtom.Data[22:24])
		mp4a[SampleRateKey] = get16Dot16FixedPointAsFloat(mp4aAtom.Data[24:28])
		if version == 1 {
			if len(mp4aAtom.Data) < 44 {
				return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 44, len(mp4aAtom.Data))
			}
			childOffset = 44
			mp4a[SamplesPerPacketKey] = getUint32AsInt64(mp4aAtom.Data[28:32])
			mp4a[BytesPerPacketKey] = getUint32AsInt64(mp4aAtom.Data[32:36])
			mp4a[BytesPerFrameKey] = getUint32AsInt64(mp4aAtom.Data[36:40])
			mp4a[BytesPerSampleKey] = getUint32AsInt64(mp4aAtom.Data[40:44])
		}
	} else if version == 2 {
		//If version is 2 then the layout of the sample description is different
		if len(mp4aAtom.Data) < 64 {
			return nil, nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 64, len(mp4aAtom.Data))
		}
		childOffset = 64
		mp4a[Always3Key] = getInt16AsInt(mp4aAtom.Data[16:18])
		mp4a[Always16Key] = getInt16AsInt(mp4aAtom.Data[18:20])
		mp4a[AlwaysMinus2Key] = getInt16AsInt(mp4aAtom.Data[20:22])
		mp4a[Always0Key] = getInt16AsInt(mp4aAtom.Data[22:24])
		mp4a[Always65536Key] = getInt32AsInt(mp4aAtom.Data[24:28])
		mp4a[SizeOfStructOnlyKey] = getInt32AsInt(mp4aAtom.Data[28:32])
		mp4a[SampleRateKey] = getFloat64(mp4aAtom.Data[32:40])
		mp4a[ChannelsKey] = getInt32AsInt(mp4aAtom.Data[40:44])
		mp4a[Always7F000000Key] = getInt32AsInt(mp4aAtom.Data[44:48])
		mp4a[ConstBitsPerChannelKey] = getInt32AsInt(mp4aAtom.Data[48:52])
		mp4a[LPCMFlagsKey] = getInt32AsInt(mp4aAtom.Data[52:56])
		mp4a[ConstBytesPerPacket] = getUint32AsInt64(mp4aAtom.Data[56:60])
		mp4a[ConstFramesPerPacket] = getUint32AsInt64(mp4aAtom.Data[60:64])
	} else {
		return nil, nil, errors.New("Unknown version.")
	}
	childReader := bytes.NewReader(mp4aAtom.Data[childOffset:])
	children, err := readMp4AtomsFunc(childReader, true, 0, false)
	if err != nil {
		return mp4a, nil, err
	}
	esdsAtom := findAtom(children, "esds")
	if esdsAtom == nil {
		return mp4a, nil, nil
	}
	esds, err := processESDSAtom(*esdsAtom)
	if err != nil {
		return mp4a, nil, err
	}
	return mp4a, esds, nil
}

type mp4esds map[string]interface{}

func processESDSAtom(esdsAtom Mp4Atom) (mp4esds, error) {
	if len(esdsAtom.Data) < 29 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29, len(esdsAtom.Data))
	}
	esds := mp4esds{}
	s := 0 //extended tag size
	esds[VersionKey] = getUint32AsInt64(esdsAtom.Data[0:4])
	esds["esDescriptorTypeTag"] = esdsAtom.Data[4]
	if esdsAtom.Data[5] == 0x80 && esdsAtom.Data[6] == 0x80 && esdsAtom.Data[7] == 0x80 {
		s = 3
	}
	if len(esdsAtom.Data) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(esdsAtom.Data))
	}
	esds["esDescriptorLength"] = esdsAtom.Data[5+s]
	esds["esID"] = getUint16AsInt(esdsAtom.Data[6+s : 8+s])
	esds["streamPriority"] = esdsAtom.Data[8+s]
	esds["decoderConfigDescriptorTypeTag"] = esdsAtom.Data[9+s]
	if esdsAtom.Data[10+s] == 0x80 && esdsAtom.Data[11+s] == 0x80 && esdsAtom.Data[12+s] == 0x80 {
		s += 3
	}
	if len(esdsAtom.Data) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(esdsAtom.Data))
	}
	esds["decoderConfigDescriptorLength"] = esdsAtom.Data[10+s]
	esds["objectTypeID"] = esdsAtom.Data[11+s]
	esds["streamType"] = esdsAtom.Data[12+s] >> 2
	esds["upstream"] = getBit(esdsAtom.Data[12+s], 1)
	esds["reserved"] = getBit(esdsAtom.Data[12+s], 0)
	esds["bufferSize"] = getUint24AsInt(esdsAtom.Data[13+s : 16+s])
	esds[MaximumBitrateKey] = getUint32AsInt64(esdsAtom.Data[16+s : 20+s])
	esds[AverageBitrateKey] = getUint32AsInt64(esdsAtom.Data[20+s : 24+s])
	esds["decoderSpecificDescriptorTypeTag"] = esdsAtom.Data[24+s]
	if esdsAtom.Data[25+s] == 0x80 && esdsAtom.Data[26+s] == 0x80 && esdsAtom.Data[27+s] == 0x80 {
		s += 3
	}
	if len(esdsAtom.Data) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(esdsAtom.Data))
	}
	esds["decoderSpecificDescriptorLength"] = esdsAtom.Data[25+s]
	esds["decoderSpecificInfo"] = esdsAtom.Data[26+s : 26+s+int(esds["decoderSpecificDescriptorLength"].(uint8))]
	s += int(esds["decoderSpecificDescriptorLength"].(uint8))
	if len(esdsAtom.Data) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(esdsAtom.Data))
	}
	esds["slConfigDescriptorTypeTag"] = esdsAtom.Data[26+s]
	if esdsAtom.Data[27+s] == 0x80 && esdsAtom.Data[28+s] == 0x80 && esdsAtom.Data[29+s] == 0x80 {
		s += 3
	}
	if len(esdsAtom.Data) < 29+s {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 29+s, len(esdsAtom.Data))
	}
	esds["slConfigDescriptorLength"] = esdsAtom.Data[27+s]
	esds["slValue"] = esdsAtom.Data[28+s]
	return esds, nil
}

func (m mp4esds) AverageBitrate() int {
	b, ok := m[AverageBitrateKey].(int64)
	if !ok || b == 0 {
		b, _ = m[MaximumBitrateKey].(int64)
	}
	//Although bitrates are stored as uint32 in an esds atom and could
	//theoretically exceed an int32, such bitrates would be unreasonable.
	//Converting down to int here is very low risk for the sake of type sanity.
	return int(b)
}
