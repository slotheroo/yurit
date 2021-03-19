package yurit

import (
	"io"
)

type mp3XingHeader map[string]interface{}

//readXingHeader reads reads information in a Xing or Info header. The method
//requires that the reader has been positioned correctly BEFORE the method is
//called.
func readMP3XingHeader(r io.Reader) (mp3XingHeader, error) {
	x := mp3XingHeader{}
	//xingIntro = ID and flags
	xingIntro, err := readBytes(r, 8)
	if err != nil {
		return nil, err
	}
	//First four bytes are the ID ("Xing" or "Info")
	x["id"] = string(xingIntro[:4])
	hasFrames := xingIntro[7]&0x01 == 1
	hasBytes := (xingIntro[7]>>1)&0x01 == 1
	hasTOC := (xingIntro[7]>>2)&0x01 == 1
	hasQuality := (xingIntro[7]>>3)&0x01 == 1
	var (
		readSize = 0
		offset   = 0
	)
	if hasFrames {
		readSize += 4
	}
	if hasBytes {
		readSize += 4
	}
	if hasTOC {
		readSize += 100
	}
	if hasQuality {
		readSize += 4
	}
	xingData, err := readBytes(r, uint(readSize))
	if err != nil {
		return nil, err
	}
	if hasFrames {
		x[TotalFramesKey] = getInt32AsInt(xingData[offset : offset+4])
		offset += 4
	}
	if hasBytes {
		x[TotalBytesKey] = getInt32AsInt(xingData[offset : offset+4])
		offset += 4
	}
	if hasTOC {
		x["toc"] = xingData[offset : offset+100]
		offset += 100
	}
	if hasQuality {
		x["qualityIndicator"] = getInt32AsInt(xingData[offset : offset+4])
		offset += 4
	}
	return x, nil
}

func (x mp3XingHeader) ID() string {
	i, _ := x["id"].(string)
	return i
}

func (x mp3XingHeader) TOC() []byte {
	if t, ok := x["toc"].([]byte); ok {
		return t
	}
	return nil
}

func (x mp3XingHeader) TotalBytes() *int {
	if t, ok := x[TotalBytesKey].(int); ok {
		return &t
	}
	return nil
}

func (x mp3XingHeader) TotalFrames() *int {
	if t, ok := x[TotalFramesKey].(int); ok {
		return &t
	}
	return nil
}

func (x mp3XingHeader) Quality() *int {
	if t, ok := x["qualityIndicator"].(int); ok {
		return &t
	}
	return nil
}
