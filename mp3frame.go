package yurit

/*
//MP3FrameData represents additional frame information that is presented after
//the frame header. This type only applies to layer III mpeg files (mp3s). Side
//information is always expected whereas the CRC bytes and the Xing header may
//or may not be present.
type MP3FrameData struct {
	CRC             []byte
	SideInformation []byte
	XingHeader      *XingHeader
}

//readMP3FrameData reads specific elements of non-audio frame data. The method
//requires that the readseeker has been positioned correctly BEFORE the method
//is called.
func readMPEGFrameData(r io.ReadSeeker, fh mpegFrameHeader) (MP3FrameData, error) {
	var (
		err          error
		mp3FrameData MP3FrameData
	)
	//This method does not read proper data for anything except an MP3, exit if else
	if fh.Layer() != MPEGLayer3 {
		return mp3FrameData, nil
	}
	//If protected is true, expect 2 bytes of CRC data
	if fh.Protected() {
		mp3FrameData.CRC, err = readBytes(r, 2)
		if err != nil {
			return mp3FrameData, err
		}
	}
	//Amount of side information depends on version and channel mode
	mp3FrameData.SideInformation, err = readBytes(r, mp3SideInfoByteLength(fh.Version(), fh.ChannelMode()))
	if err != nil {
		return mp3FrameData, err
	}
	//After the side information is where the Xing header may appear
	xingCheck, err := readBytes(r, 4)
	if err != nil {
		return mp3FrameData, err
	}
	//Expected xing location, if it exists
	if string(xingCheck) == XING_XING_ID || string(xingCheck) == XING_INFO_ID {
		_, err := r.Seek(-4, io.SeekCurrent)
		x, err := readXingHeader(r)
		mp3FrameData.XingHeader = &x
		if err != nil {
			return mp3FrameData, err
		}
	} else if fh.Protected() && len(mp3FrameData.SideInformation) >= 2 {
		//Look for xing in commonly misplaced location
		xingCheck2 := mp3FrameData.SideInformation[len(mp3FrameData.SideInformation)-2:]
		xingCheck2 = append(xingCheck2, xingCheck[:2]...)
		if string(xingCheck2) == XING_XING_ID || string(xingCheck2) == XING_INFO_ID {
			_, err := r.Seek(-6, io.SeekCurrent)
			x, err := readXingHeader(r)
			mp3FrameData.XingHeader = &x
			if err != nil {
				return mp3FrameData, err
			}
		}
	}
	return mp3FrameData, nil
}

//readMP3FrameData reads specific elements of non-audio frame data. The method
//requires that the readseeker has been positioned correctly BEFORE the method
//is called.
func readMP3FrameData(r io.ReadSeeker, fh mpegFrameHeader) (MP3FrameData, error) {
	var (
		err          error
		mp3FrameData MP3FrameData
	)
	//This method does not read proper data for anything except an MP3, exit if else
	if fh.Layer() != MPEGLayer3 {
		return mp3FrameData, nil
	}
	//If protected is true, expect 2 bytes of CRC data
	if fh.Protected() {
		mp3FrameData.CRC, err = readBytes(r, 2)
		if err != nil {
			return mp3FrameData, err
		}
	}
	//Amount of side information depends on version and channel mode
	mp3FrameData.SideInformation, err = readBytes(r, mp3SideInfoByteLength(fh.Version(), fh.ChannelMode()))
	if err != nil {
		return mp3FrameData, err
	}
	//After the side information is where the Xing header may appear
	xingCheck, err := readBytes(r, 4)
	if err != nil {
		return mp3FrameData, err
	}
	//Expected xing location, if it exists
	if string(xingCheck) == XING_XING_ID || string(xingCheck) == XING_INFO_ID {
		_, err := r.Seek(-4, io.SeekCurrent)
		x, err := readXingHeader(r)
		mp3FrameData.XingHeader = &x
		if err != nil {
			return mp3FrameData, err
		}
	} else if fh.Protected() && len(mp3FrameData.SideInformation) >= 2 {
		//Look for xing in commonly misplaced location
		xingCheck2 := mp3FrameData.SideInformation[len(mp3FrameData.SideInformation)-2:]
		xingCheck2 = append(xingCheck2, xingCheck[:2]...)
		if string(xingCheck2) == XING_XING_ID || string(xingCheck2) == XING_INFO_ID {
			_, err := r.Seek(-6, io.SeekCurrent)
			x, err := readXingHeader(r)
			mp3FrameData.XingHeader = &x
			if err != nil {
				return mp3FrameData, err
			}
		}
	}
	return mp3FrameData, nil
}

//Returns the length of the side information based on version and channel mode.
func mp3SideInfoByteLength(version MPEGVersion, channelMode MPEGChannelMode) uint {
	//Don't bother with reserved
	//version 1 && non-mono 32B (else it's 17)
	//version 2/2.5 mono is 9B (else it's 17)
	if version == MPEGVersionReserved {
		return 0
	} else if version == MPEGVersion_1 && channelMode != MPEGChannelSingle {
		return 32
	} else if (version == MPEGVersion_2 || version == MPEGVersion_2_5) && channelMode == MPEGChannelSingle {
		return 9
	}
	return 17
}

//The XingHeader is an optional header found in within the data section of an
//MP3 frame. Note that the scope of this type includes both a header with the
//Xing ID as well as a header with the Info ID. Typically Xing represents a
//variable bitrate file and Info represents a constant bitrate file. Xing
//information is helpful for calculating mp3 duration in a variable bitrate file
//without needing to read every frame.
type XingHeader struct {
	ID      string
	Frames  *int
	Bytes   *int
	TOC     []byte
	Quality *int
}

//All possible version values.
const (
	XING_XING_ID string = "Xing"
	XING_INFO_ID string = "Info"
)

//readXingHeader reads reads information in a Xing or Info header. The method
//requires that the reader has been positioned correctly BEFORE the method is
//called.
func readXingHeader(r io.Reader) (XingHeader, error) {
	var (
		xingHeader XingHeader
	)
	//xingIntro = ID and flags
	xingIntro, err := readBytes(r, 8)
	if err != nil {
		return xingHeader, err
	}
	//First four bytes are the ID ("Xing" or "Info")
	xingHeader.ID = string(xingIntro[:4])
	//If Frames flag is set, read number of frames
	if getBit(xingIntro[7], 0) {
		numFramesBytes, err := readBytes(r, 4)
		if err != nil {
			return xingHeader, err
		}
		numFrames := getInt(numFramesBytes)
		xingHeader.Frames = &numFrames
	}
	//If Bytes flag is set, read number of bytes
	if getBit(xingIntro[7], 1) {
		numBytesBytes, err := readBytes(r, 4)
		if err != nil {
			return xingHeader, err
		}
		numBytes := getInt(numBytesBytes)
		xingHeader.Bytes = &numBytes
	}
	//If TOC flag is set, read TOC bytes
	if getBit(xingIntro[7], 2) {
		xingHeader.TOC, err = readBytes(r, 100)
		if err != nil {
			return xingHeader, err
		}
	}
	//If Quality flag is set, read quality indicator
	if getBit(xingIntro[7], 3) {
		qualityBytes, err := readBytes(r, 4)
		if err != nil {
			return xingHeader, err
		}
		quality := getInt(qualityBytes)
		xingHeader.Quality = &quality
	}
	return xingHeader, nil
}
*/
