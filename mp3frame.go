package yurit

import (
	"io"
)

//MP3FrameHeader represents the information contained in the first frame header
//encountered in an mp3 file. Much of this information would be consistent for
//all frame headers in a file, but some would vary from frame to frame. For
//instance, the version and layer will be the same in all frames, but the
//bitrate could vary from frame to frame in a variable bitrate (VBR) file.
//Technically this frame header can apply to different MPEG layers (e.g. mp2) so
//the name MP3FrameHeader is not entirely apt, but an exhaustively descriptive
//name would be cumbersome.
type MP3FrameHeader struct {
	Version       MPEGVersion
	Layer         MPEGLayer
	Protected     bool //True for 0 bit! Indicates that a CRC follows the header.
	Bitrate       int  //Frame bitrate in kilobits per second (e.g. 128 = 128000bps)
	SamplingRate  int  //File sampling rate frequency in hertz
	Padded        bool //True for 1 bit. Indicates that this frame is padded with one slot.
	Private       bool //True for 1 bit. So private that no one knows what this is for.
	ChannelMode   MPEGChannelMode
	ModeExtension MPEGModeExtension
	Copyright     bool //True for 1 bit. Indicates that the audio is copyrighted.
	Original      bool //True for 1 bit. Indicates that this is the original media.
	Emphasis      MPEGEmphasis
}

//MPEGVersion is the audio version ID for the file. For most common MP3 files
//this will almost always be MPEG Version 1. The sampling rate for a file will
//exclusively map to one of these versions. (e.g. All 44.1 kHz files are MPEG
//Version 1)
type MPEGVersion string

//All possible version values.
const (
	MPEGVersion_2_5     MPEGVersion = "MPEG Version 2.5"
	MPEGVersionReserved MPEGVersion = "reserved"
	MPEGVersion_2       MPEGVersion = "MPEG Version 2"
	MPEGVersion_1       MPEGVersion = "MPEG Version 1"
)

//maps header byte to version value
var mpegVersionMap = map[byte]MPEGVersion{
	0: MPEGVersion_2_5,
	1: MPEGVersionReserved,
	2: MPEGVersion_2,
	3: MPEGVersion_1,
}

//MPEGLayer is the layer index for the file. For an MP3 this will be Layer III,
//an MP2 would be Layer II, and an MP1 would be Layer I.
type MPEGLayer string

//All possible layer values
const (
	MPEGLayerReserved MPEGLayer = "reserved"
	MPEGLayer3        MPEGLayer = "Layer III" //mp3
	MPEGLayer2        MPEGLayer = "Layer II"  //mp2
	MPEGLayer1        MPEGLayer = "Layer I"   //mp1
)

//maps header byte to layer value
var mpegLayerMap = map[byte]MPEGLayer{
	0: MPEGLayerReserved,
	1: MPEGLayer3,
	2: MPEGLayer2,
	3: MPEGLayer1,
}

//If the version and layer data in the header maps to an invalid bitrate then a
//value of -1 is returned for the bitrate. If the header data maps to a free
//bitrate then a value of 0 is returned for the bitrate.
const (
	BadBitrate  = -1
	FreeBitrate = 0
)

var badBitrateSlice = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

//maps version, layer, and byte value to the correct bitrate
var mpegBitrateMap = map[MPEGVersion]map[MPEGLayer][]int{
	MPEGVersionReserved: map[MPEGLayer][]int{
		MPEGLayerReserved: badBitrateSlice,
		MPEGLayer1:        badBitrateSlice,
		MPEGLayer2:        badBitrateSlice,
		MPEGLayer3:        badBitrateSlice,
	},
	MPEGVersion_1: map[MPEGLayer][]int{
		MPEGLayerReserved: badBitrateSlice,
		MPEGLayer1:        []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, -1},
		MPEGLayer2:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, -1},
		MPEGLayer3:        []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, -1},
	},
	MPEGVersion_2: map[MPEGLayer][]int{
		MPEGLayerReserved: badBitrateSlice,
		MPEGLayer1:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, -1},
		MPEGLayer2:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
		MPEGLayer3:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
	},
	MPEGVersion_2_5: map[MPEGLayer][]int{
		MPEGLayerReserved: badBitrateSlice,
		MPEGLayer1:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, -1},
		MPEGLayer2:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
		MPEGLayer3:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1},
	},
}

//If either the version or the layer has a value of reserved then the sampling
//rate will be returned as -1. Any valid combination of version and layer will
//result in a valid sampling rate value.
const BadSamplingRate = -1

//maps version and byte value to the correct sampling rate
var mpegSamplingRateMap = map[MPEGVersion][]int{
	MPEGVersionReserved: []int{-1, -1, -1, -1},
	MPEGVersion_1:       []int{44100, 48000, 32000, -1},
	MPEGVersion_2:       []int{22050, 24000, 16000, -1},
	MPEGVersion_2_5:     []int{11025, 12000, 8000, -1},
}

//MPEGChannelMode is simply the channel mode for the audio
type MPEGChannelMode string

//All possible channel mode values
const (
	MPEGChannelStereo      MPEGChannelMode = "Stereo"
	MPEGChannelJointStereo MPEGChannelMode = "Joint stereo (Stereo)"
	MPEGChannelDual        MPEGChannelMode = "Dual channel (2 mono channels)"
	MPEGChannelSingle      MPEGChannelMode = "Single channel (Mono)"
)

//maps the header byte to the corresponding channel mode
var mpegChannelModeMap = map[byte]MPEGChannelMode{
	0: MPEGChannelStereo,
	1: MPEGChannelJointStereo,
	2: MPEGChannelDual,
	3: MPEGChannelSingle,
}

//MPEGModeExtension provides additional information about how the audio is
//encoded if the channel mode is 'Joint stereo'. The mode extension is not
//applicable for other channel modes.
type MPEGModeExtension string

const (
	//Channel mode is not joint stereo
	MPEGModeExtensionNA MPEGModeExtension = "not applicable"
	//These four mode extensions only apply to layers I and II
	MPEGModeExtension4To31  MPEGModeExtension = "bands 4 to 31"
	MPEGModeExtension8To31  MPEGModeExtension = "bands 8 to 31"
	MPEGModeExtension12To31 MPEGModeExtension = "bands 12 to 31"
	MPEGModeExtension16To31 MPEGModeExtension = "bands 16 to 31"
	//These four mode extensions only apply to layer III
	MPEGModeExtensionMSOffIntensityOff MPEGModeExtension = "M/S stereo off, Intensity stereo off"
	MPEGModeExtensionMSOffIntensityOn  MPEGModeExtension = "M/S stereo off, Intensity stereo on"
	MPEGModeExtensionMSOnIntensityOff  MPEGModeExtension = "M/S stereo on, Intensity stereo off"
	MPEGModeExtensionMSOnIntensityOn   MPEGModeExtension = "M/S stereo on, Intensity stereo on"
)

//maps layer and byte value to a mode extension (assuming channel mode is Joint stereo)
var mpegModeExtensionMap = map[MPEGLayer][]MPEGModeExtension{
	MPEGLayerReserved: []MPEGModeExtension{MPEGModeExtensionNA, MPEGModeExtensionNA, MPEGModeExtensionNA, MPEGModeExtensionNA},
	MPEGLayer3:        []MPEGModeExtension{MPEGModeExtensionMSOffIntensityOff, MPEGModeExtensionMSOffIntensityOn, MPEGModeExtensionMSOnIntensityOff, MPEGModeExtensionMSOnIntensityOn},
	MPEGLayer2:        []MPEGModeExtension{MPEGModeExtension4To31, MPEGModeExtension8To31, MPEGModeExtension12To31, MPEGModeExtension16To31},
	MPEGLayer1:        []MPEGModeExtension{MPEGModeExtension4To31, MPEGModeExtension8To31, MPEGModeExtension12To31, MPEGModeExtension16To31},
}

//MPEGEmphasis gives the decoder instructions on how to de-emphasize sound in
//the file. It is rarely used.
type MPEGEmphasis string

//All possible emphasis values
const (
	MPEGEmphasisNone     MPEGEmphasis = "none"
	MPEGEmphasis50_15    MPEGEmphasis = "50/15 ms"
	MPEGEmphasisReserved MPEGEmphasis = "reserved"
	MPEGEmphasisCCIT     MPEGEmphasis = "CCIT J.17"
)

//maps the byte in the header to the corresponding emphasis value
var mpegEmphasisMap = map[byte]MPEGEmphasis{
	0: MPEGEmphasisNone,
	1: MPEGEmphasis50_15,
	2: MPEGEmphasisReserved,
	3: MPEGEmphasisCCIT,
}

//readMP3FrameHeader searches for the frame sync match (11111111 111xxxxx) then
//reads the first frame header encountered
//Bits: AAAAAAAA AAABBCCD EEEEFFGH IIJJKLMM
//A = Frame sync (all 1s), B = MPEG audio version, C = Layer description
//D = Protection bit, E = Bitrate index, F = Sampling rate index,
//G = Padding bit, H = Private bit, I = Channel mode, J = Mode extension,
//K = Copyright, L = Original, M = Emphasis
func readMP3FrameHeader(r io.Reader) (MP3FrameHeader, error) {
	var (
		numBytesToRead uint = 4
		buff           []byte
		mp3FrameHeader MP3FrameHeader
	)
	//Read bytes until we find a frame sync match
	for {
		b, err := readBytes(r, numBytesToRead)
		if err != nil {
			return mp3FrameHeader, err
		}
		//This is always expected to fill buff to exactly 4 bytes
		buff = append(buff, b...)
		//If frame sync match, break. Else see if the match is or may be  present
		//but is not yet aligned to the first byte.
		if buff[0] == 0xFF && (buff[1]&0xE0 == 0xE0) {
			break
		} else if buff[1] == 0xFF && (buff[2]&0xE0 == 0xE0) {
			numBytesToRead = 1
			buff = buff[1:]
		} else if buff[2] == 0xFF && (buff[3]&0xE0 == 0xE0) {
			numBytesToRead = 2
			buff = buff[2:]
		} else if buff[3] == 0xFF {
			numBytesToRead = 3
			buff = buff[3:]
		} else {
			numBytesToRead = 4
			buff = []byte{}
		}
	}
	mp3FrameHeader.Version = extractMPEGVersion(buff[1])
	mp3FrameHeader.Layer = extractMPEGLayer(buff[1])
	mp3FrameHeader.Protected = !getBit(buff[1], 0) //The 1 bit means NOT protected
	mp3FrameHeader.Bitrate = extractMPEGBitrate(buff[2], mp3FrameHeader.Version, mp3FrameHeader.Layer)
	mp3FrameHeader.SamplingRate = extractMPEGSamplingRate(buff[2], mp3FrameHeader.Version)
	mp3FrameHeader.Padded = getBit(buff[2], 1)
	mp3FrameHeader.Private = getBit(buff[2], 0)
	mp3FrameHeader.ChannelMode = extractMPEGChannelMode(buff[3])
	if mp3FrameHeader.ChannelMode == MPEGChannelJointStereo {
		mp3FrameHeader.ModeExtension = extractMPEGModeExtension(buff[3], mp3FrameHeader.Layer)
	} else {
		mp3FrameHeader.ModeExtension = MPEGModeExtensionNA
	}
	mp3FrameHeader.Copyright = getBit(buff[3], 3)
	mp3FrameHeader.Original = getBit(buff[3], 2)
	mp3FrameHeader.Emphasis = extractMPEGEmphasis(buff[3])
	return mp3FrameHeader, nil
}

//Given the second byte of a frame header, returns the version.
func extractMPEGVersion(b byte) MPEGVersion {
	key := (b >> 3) & 0x03
	return mpegVersionMap[key]
}

//Given the second byte of a frame header, returns the layer.
func extractMPEGLayer(b byte) MPEGLayer {
	key := (b >> 1) & 0x03
	return mpegLayerMap[key]
}

//Given the third byte of a frame header and other data, returns the bitrate.
func extractMPEGBitrate(b byte, version MPEGVersion, layer MPEGLayer) int {
	key := int((b >> 4) & 0x0F)
	return mpegBitrateMap[version][layer][key]
}

//Given the third byte of a frame header and the version, returns the sampling rate.
func extractMPEGSamplingRate(b byte, version MPEGVersion) int {
	key := int((b >> 2) & 0x03)
	return mpegSamplingRateMap[version][key]
}

//Given the fourth byte of a frame header, returns the channel mode.
func extractMPEGChannelMode(b byte) MPEGChannelMode {
	key := (b >> 6) & 0x03
	return mpegChannelModeMap[key]
}

//Given the fourth byte of a frame header and the layer, returns the mode extension.
//This method assumes that the channel mode has already been esetablished as
//Joint stereo.
func extractMPEGModeExtension(b byte, layer MPEGLayer) MPEGModeExtension {
	key := int((b >> 4) & 0x03)
	return mpegModeExtensionMap[layer][key]
}

//Given the fourth byte of a frame header, returns the emphasis.
func extractMPEGEmphasis(b byte) MPEGEmphasis {
	key := b & 0x03
	return mpegEmphasisMap[key]
}

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
func readMP3FrameData(r io.ReadSeeker, h MP3FrameHeader) (MP3FrameData, error) {
	var (
		err          error
		mp3FrameData MP3FrameData
	)
	//This method does not read proper data for anything except an MP3, exit if else
	if h.Layer != MPEGLayer3 {
		return mp3FrameData, nil
	}
	//If protected is true, expect 2 bytes of CRC data
	if h.Protected {
		mp3FrameData.CRC, err = readBytes(r, 2)
		if err != nil {
			return mp3FrameData, err
		}
	}
	//Amount of side information depends on version and channel mode
	mp3FrameData.SideInformation, err = readBytes(r, mp3SideInfoByteLength(h.Version, h.ChannelMode))
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
	} else if h.Protected && len(mp3FrameData.SideInformation) >= 2 {
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