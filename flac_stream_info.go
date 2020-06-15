package yurit

import (
	"time"
)

type flacStreamInfo map[string]interface{}

//readStreamInfoBlock reads the STREAMINFO block from a FLAC file
//https://xiph.org/flac/format.html#metadata_block_streaminfo
func processStreamInfoBlock(b []byte) (flacStreamInfo, error) {
	if err := checkLen(b, 34); err != nil {
		return nil, err
	}
	si := flacStreamInfo{}
	si[MinimumBlockSizeKey] = getUint16AsInt(b[0:2])
	si[MaximumBlockSizeKey] = getUint16AsInt(b[2:4])
	si[MinimumFrameSizeKey] = getUint24AsInt(b[4:7])
	si[MaximumFrameSizeKey] = getUint24AsInt(b[7:10])
	si[SampleRateKey] = getUint24AsInt(b[10:13]) >> 4
	si[ChannelsKey] = ((b[12] >> 1) & 0x07) + 1
	si[SampleSizeKey] = ((b[12] & 0x01) << 4) + (b[13] >> 4) + 1
	si[TotalSamplesKey] = getInt64(append([]byte{0, 0, 0, b[13] & 0x0F}, b[14:18]...))
	si[MD5Key] = b[18:]
	return si, nil
}

func (si flacStreamInfo) Duration() time.Duration {
	sr, srok := si[SampleRateKey].(int)
	ts, tsok := si[TotalSamplesKey].(int64)
	if srok && tsok && sr != 0 {
		//Calculate track length by dividing total samples by sample rate
		seconds := float64(ts) / float64(sr)
		//convert to time.Duration
		return time.Duration(seconds * float64(time.Second))
	}
	return time.Duration(0)
}

func (si flacStreamInfo) SampleRate() int {
	sr, _ := si[SampleRateKey].(int)
	return sr
}
