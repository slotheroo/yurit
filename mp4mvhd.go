package yurit

import (
	"fmt"
	"time"
)

type mp4mvhd map[string]interface{}

func processMVHDAtom(mvhdAtom Mp4Atom) (mp4mvhd, error) {
	if len(mvhdAtom.Data) < 100 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 100, len(mvhdAtom.Data))
	}
	m := mp4mvhd{}
	m[VersionKey] = mvhdAtom.Data[0]
	m[FlagsKey] = mvhdAtom.Data[1:4]
	m["creationTime"] = getUint32AsInt64(mvhdAtom.Data[4:8])
	m["modificationTime"] = getUint32AsInt64(mvhdAtom.Data[8:12])
	m[TimeScaleKey] = getUint32AsInt64(mvhdAtom.Data[12:16])
	m[DurationKey] = getUint32AsInt64(mvhdAtom.Data[16:20])
	m["preferredRate"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[20:24])
	m["preferredVolume"] = get8Dot8FixedPointAsFloat(mvhdAtom.Data[24:26])
	//Skip 10 bytes reserved
	m["matrixA"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[36:40])
	m["matrixB"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[40:44])
	m["matrixU"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[44:48])
	m["matrixC"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[48:52])
	m["matrixD"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[52:56])
	m["matrixV"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[56:60])
	m["matrixX"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[60:64])
	m["matrixY"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[64:68])
	m["matrixW"] = get16Dot16FixedPointAsFloat(mvhdAtom.Data[68:72])
	m["previewTime"] = getUint32AsInt64(mvhdAtom.Data[72:76])
	m["previewDuration"] = getUint32AsInt64(mvhdAtom.Data[76:80])
	m["posterTime"] = getUint32AsInt64(mvhdAtom.Data[80:84])
	m["selectionTime"] = getUint32AsInt64(mvhdAtom.Data[84:88])
	m["selectionDuration"] = getUint32AsInt64(mvhdAtom.Data[88:92])
	m["currentTime"] = getUint32AsInt64(mvhdAtom.Data[92:96])
	m["nextTrackID"] = getInt32AsInt(mvhdAtom.Data[96:100])
	return m, nil
}

func (m mp4mvhd) Duration() time.Duration {
	d, okd := m[DurationKey].(int64)
	t, okt := m[TimeScaleKey].(int64)
	if okd && okt && t != 0 {
		//Calculate true duration by dividing duration (total samples) by time scale (sample rate)
		seconds := float64(d) / float64(t)
		return time.Duration(seconds * float64(time.Second))
	}
	//Missing data, cannot calculate, return 0 as a duration
	return time.Duration(0)
}
