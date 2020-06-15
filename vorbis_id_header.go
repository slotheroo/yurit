package yurit

import (
	"encoding/binary"
	"time"
)

//VorbisIDHeader holds general information about a Vorbis audio stream.
//https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-630004.2.2
type vorbisIDHeader map[string]interface{}

//Reads the identification header from a Vorbis audio stream
//See https://xiph.org/vorbis/doc/Vorbis_I_spec.html#x1-630004.2.2
func processVorbisIDHeader(b []byte) (vorbisIDHeader, error) {
	vih := vorbisIDHeader{}
	//Identification header is 23 bytes long
	if err := checkLen(b, 23); err != nil {
		return nil, err
	}
	vih[VersionKey] = getUint32LittleAsInt64(b[0:4])
	vih[ChannelsKey] = b[4]
	vih[SampleRateKey] = getUint32LittleAsInt64(b[5:9])
	vih[MaximumBitrateKey] = getInt32LittleAsInt(b[9:13])
	vih[AverageBitrateKey] = getInt32LittleAsInt(b[13:17]) //aka nominal bitrate
	vih[MinimumBitrateKey] = getInt32LittleAsInt(b[17:21])
	//Use bits 0-3 of byte 21 to make a uint and use that as an exponent of 2
	vih[MinimumBlockSizeKey] = 1 << binary.LittleEndian.Uint16([]byte{b[21] & 0x0F, 0})
	//Use bits 4-7 of byte 21 to make a uint and use that as an exponent of 2
	vih[MaximumBlockSizeKey] = 1 << binary.LittleEndian.Uint16([]byte{b[21] >> 4, 0})
	//Last byte is the framing flag.
	vih["framing"] = b[22]
	return vih, nil
}

func (vih vorbisIDHeader) AverageBitrate() int {
	// Bitrate nominal represents the average bitrate if it is set
	b, _ := vih[AverageBitrateKey].(int)
	if b == 0 {
		// Bitrate nominal not set so we'll try to guess the average
		bmax, _ := vih[MaximumBitrateKey].(int)
		bmin, _ := vih[MinimumBitrateKey].(int)
		if bmax != 0 && bmin != 0 {
			// Bitrate nominal is not set, but max and min are. Take the average
			b = (bmax + bmin) / 2
		} else if bmax != 0 {
			// Bitrate max is the only value set, return it
			b = bmax
		} else {
			// Else return bitrate min, even if it is 0
			b = bmin
		}
	}
	return b
}

func (vih vorbisIDHeader) Duration(totalGranules int64) time.Duration {
	sr, ok := vih[SampleRateKey].(int64)
	if ok && sr != 0 {
		//Calculate track length by dividing total samples (granules) by sample rate
		seconds := float64(totalGranules) / float64(sr)
		//convert to time.Duration
		return time.Duration(seconds * float64(time.Second))
	}
	//Missing data, cannot calculate, return 0 as a duration
	return time.Duration(0)
}
