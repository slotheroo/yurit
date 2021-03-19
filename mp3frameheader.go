package yurit

import (
	"io"
)

//mpegFrameHeader represents the information contained in the first frame header
//encountered in an mp3 file. Much of this information would be consistent for
//all frame headers in a file, but some would vary from frame to frame. For
//instance, the version and layer will be the same in all frames, but the
//bitrate could vary from frame to frame in a variable bitrate (VBR) file.
type mpegFrameHeader map[string]interface{}

//readMPEGFrameHeader searches for the frame sync match (11111111 111xxxxx) then
//reads the first frame header encountered
//Bits: AAAAAAAA AAABBCCD EEEEFFGH IIJJKLMM
//A = Frame sync (all 1s), B = MPEG audio version, C = Layer description
//D = Protection bit, E = Bitrate index, F = Sampling rate index,
//G = Padding bit, H = Private bit, I = Channel mode, J = Mode extension,
//K = Copyright, L = Original, M = Emphasis
func readMPEGFrameHeader(r io.Reader) (mpegFrameHeader, error) {
	var (
		numBytesToRead uint = 4
		buff           []byte
		fh             = mpegFrameHeader{}
	)
	//Read bytes until we find a frame sync match
	for {
		b, err := readBytes(r, numBytesToRead)
		if err != nil {
			return fh, err
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
	fh[VersionKey] = (buff[1] >> 3) & 0x03      //AAABB>>> & 00000011
	fh["layer"] = (buff[1] >> 1) & 0x03         //AAABBCC> & 00000011
	fh["protected"] = buff[1] & 0x01            //AAABBCCD & 00000001
	fh["bitrate"] = (buff[2] >> 4) & 0x0F       //EEEE>>>> & 00001111
	fh[SampleRateKey] = (buff[2] >> 2) & 0x03   //EEEEFF>> & 00000011
	fh["padded"] = (buff[2] >> 1) & 0x01        //EEEEFFG> & 00000001
	fh["private"] = buff[2] & 0x01              //EEEEFFGH & 00000001
	fh[ChannelsKey] = (buff[3] >> 6) & 0x03     //II>>>>>> & 00000011
	fh["modeExtension"] = (buff[3] >> 4) & 0x03 //IIJJ>>>> & 00000011
	fh["copyright"] = (buff[3] >> 3) & 0x01     //IIJJK>>> & 00000001
	fh["original"] = (buff[3] >> 2) & 0x01      //IIJJKL>> & 00000001
	fh["emphasis"] = buff[3] & 0x03             //IIJJKLMM & 00000011
	return fh, nil
}

func (fh mpegFrameHeader) Bitrate() int {
	bitrateIndex, ok := fh["bitrate"].(byte)
	if !ok {
		return 0
	}
	var mpegBitrateMap = map[MPEGVersion]map[MPEGLayer][]int{
		MPEGVersionReserved: map[MPEGLayer][]int{
			MPEGLayerReserved: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer1:        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer2:        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer3:        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		MPEGVersion_1: map[MPEGLayer][]int{
			MPEGLayerReserved: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer1:        []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0},
			MPEGLayer2:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 0},
			MPEGLayer3:        []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0},
		},
		MPEGVersion_2: map[MPEGLayer][]int{
			MPEGLayerReserved: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer1:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0},
			MPEGLayer2:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
			MPEGLayer3:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
		},
		MPEGVersion_2_5: map[MPEGLayer][]int{
			MPEGLayerReserved: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			MPEGLayer1:        []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0},
			MPEGLayer2:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
			MPEGLayer3:        []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
		},
	}
	bitrateSlice, ok := mpegBitrateMap[fh.Version()][fh.Layer()]
	if !ok {
		return 0
	}
	if int(bitrateIndex) >= len(bitrateSlice) {
		return 0
	}
	return bitrateSlice[bitrateIndex]
}

func (fh mpegFrameHeader) ChannelMode() MPEGChannelMode {
	cm, ok := fh[ChannelsKey].(byte)
	if !ok {
		return ""
	}
	if cm == 0 {
		return MPEGChannelStereo
	} else if cm == 1 {
		return MPEGChannelJointStereo
	} else if cm == 2 {
		return MPEGChannelDual
	} else if cm == 3 {
		return MPEGChannelSingle
	}
	return ""
}

func (fh mpegFrameHeader) Copyright() bool {
	c, _ := fh["copyright"].(byte)
	//1 bit means copyright is true, 0 means false
	return c == 1
}

func (fh mpegFrameHeader) Emphasis() MPEGEmphasis {
	e, ok := fh["emphasis"].(byte)
	if !ok {
		return ""
	}
	if e == 0 {
		return MPEGEmphasisNone
	} else if e == 1 {
		return MPEGEmphasis50_15
	} else if e == 2 {
		return MPEGEmphasisReserved
	} else if e == 3 {
		return MPEGEmphasisCCIT
	}
	return ""
}

func (fh mpegFrameHeader) Layer() MPEGLayer {
	l, ok := fh["layer"].(byte)
	if !ok {
		return ""
	}
	if l == 0 {
		return MPEGLayerReserved
	} else if l == 1 {
		return MPEGLayer3
	} else if l == 2 {
		return MPEGLayer2
	} else if l == 3 {
		return MPEGLayer1
	}
	return ""
}

func (fh mpegFrameHeader) ModeExtension() MPEGModeExtension {
	modeExtensionIndex, ok := fh["modeExtension"].(byte)
	if !ok {
		return ""
	}
	if fh.ChannelMode() != MPEGChannelJointStereo {
		return MPEGModeExtensionNA
	}
	var mpegModeExtensionMap = map[MPEGLayer][]MPEGModeExtension{
		MPEGLayerReserved: []MPEGModeExtension{MPEGModeExtensionNA, MPEGModeExtensionNA, MPEGModeExtensionNA, MPEGModeExtensionNA},
		MPEGLayer3:        []MPEGModeExtension{MPEGModeExtensionMSOffIntensityOff, MPEGModeExtensionMSOffIntensityOn, MPEGModeExtensionMSOnIntensityOff, MPEGModeExtensionMSOnIntensityOn},
		MPEGLayer2:        []MPEGModeExtension{MPEGModeExtension4To31, MPEGModeExtension8To31, MPEGModeExtension12To31, MPEGModeExtension16To31},
		MPEGLayer1:        []MPEGModeExtension{MPEGModeExtension4To31, MPEGModeExtension8To31, MPEGModeExtension12To31, MPEGModeExtension16To31},
	}
	modeExtensionSlice, ok := mpegModeExtensionMap[fh.Layer()]
	if !ok {
		return ""
	}
	if int(modeExtensionIndex) >= len(modeExtensionSlice) {
		return ""
	}
	return modeExtensionSlice[modeExtensionIndex]
}

func (fh mpegFrameHeader) Original() bool {
	o, _ := fh["original"].(byte)
	//1 bit means original is true, 0 means false
	return o == 1
}

func (fh mpegFrameHeader) Padded() bool {
	p, _ := fh["padded"].(byte)
	//1 bit means padded is true, 0 means false
	return p == 1
}

func (fh mpegFrameHeader) Private() bool {
	p, _ := fh["private"].(byte)
	//1 bit means private is true, 0 means false
	return p == 1
}

func (fh mpegFrameHeader) Protected() bool {
	p, ok := fh["protected"].(byte)
	//0 bit means protected is true, 1 means false
	return ok && p == 0
}

func (fh mpegFrameHeader) SampleRate() int {
	sampleRateIndex, ok := fh[SampleRateKey].(byte)
	if !ok {
		return 0
	}
	var mpegSampleRateMap = map[MPEGVersion][]int{
		MPEGVersionReserved: []int{0, 0, 0, 0},
		MPEGVersion_1:       []int{44100, 48000, 32000, 0},
		MPEGVersion_2:       []int{22050, 24000, 16000, 0},
		MPEGVersion_2_5:     []int{11025, 12000, 8000, 0},
	}
	sampleRateSlice, ok := mpegSampleRateMap[fh.Version()]
	if !ok {
		return 0
	}
	if int(sampleRateIndex) >= len(sampleRateSlice) {
		return 0
	}
	return sampleRateSlice[sampleRateIndex]
}

func (fh mpegFrameHeader) SamplesPerFrame() int {
	var samplesPerFrameMap = map[MPEGVersion]map[MPEGLayer]int{
		MPEGVersionReserved: map[MPEGLayer]int{
			MPEGLayerReserved: 0,
			MPEGLayer1:        0,
			MPEGLayer2:        0,
			MPEGLayer3:        0,
		},
		MPEGVersion_1: map[MPEGLayer]int{
			MPEGLayerReserved: 0,
			MPEGLayer1:        384,
			MPEGLayer2:        1152,
			MPEGLayer3:        1152,
		},
		MPEGVersion_2: map[MPEGLayer]int{
			MPEGLayerReserved: 0,
			MPEGLayer1:        384,
			MPEGLayer2:        1152,
			MPEGLayer3:        576,
		},
		MPEGVersion_2_5: map[MPEGLayer]int{
			MPEGLayerReserved: 0,
			MPEGLayer1:        384,
			MPEGLayer2:        1152,
			MPEGLayer3:        576,
		},
	}
	spf, _ := samplesPerFrameMap[fh.Version()][fh.Layer()]
	return spf
}

//Returns the length of the side information based on version and channel mode.
func (fh mpegFrameHeader) sideInfoLength() int {
	if fh.Layer() != MPEGLayer3 {
		return 0
	}
	var sideInfoLengthMap = map[MPEGVersion]map[MPEGChannelMode]int{
		MPEGVersionReserved: map[MPEGChannelMode]int{
			MPEGChannelStereo:      0,
			MPEGChannelJointStereo: 0,
			MPEGChannelDual:        0,
			MPEGChannelSingle:      0,
		},
		MPEGVersion_1: map[MPEGChannelMode]int{
			MPEGChannelStereo:      32,
			MPEGChannelJointStereo: 32,
			MPEGChannelDual:        32,
			MPEGChannelSingle:      17,
		},
		MPEGVersion_2: map[MPEGChannelMode]int{
			MPEGChannelStereo:      17,
			MPEGChannelJointStereo: 17,
			MPEGChannelDual:        17,
			MPEGChannelSingle:      9,
		},
		MPEGVersion_2_5: map[MPEGChannelMode]int{
			MPEGChannelStereo:      17,
			MPEGChannelJointStereo: 17,
			MPEGChannelDual:        17,
			MPEGChannelSingle:      9,
		},
	}
	sil, _ := sideInfoLengthMap[fh.Version()][fh.ChannelMode()]
	return sil
}

func (fh mpegFrameHeader) Version() MPEGVersion {
	v, ok := fh[VersionKey].(byte)
	if !ok {
		return ""
	}
	if v == 0 {
		return MPEGVersion_2_5
	} else if v == 1 {
		return MPEGVersionReserved
	} else if v == 2 {
		return MPEGVersion_2
	} else if v == 3 {
		return MPEGVersion_1
	}
	return ""
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
