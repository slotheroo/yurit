package yurit

import "fmt"

type mp4ftyp map[string]interface{}

func processFTYPAtom(ftypAtom Mp4Atom) (mp4ftyp, error) {
	if len(ftypAtom.Data) < 12 {
		return nil, fmt.Errorf("invalid encoding: expected at least %d bytes, got %d", 12, len(ftypAtom.Data))
	}
	m := mp4ftyp{}
	m[MajorBrandKey] = string(ftypAtom.Data[:4])
	m[MinorVersionKey] = ftypAtom.Data[4:8]
	cb := []string{string(ftypAtom.Data[8:12])}
	for i := 12; i+4 < len(ftypAtom.Data); i += 4 {
		cb = append(cb, string(ftypAtom.Data[i:i+4]))
	}
	m[CompatibleBrandsKey] = cb
	return m, nil
}

func (m mp4ftyp) FileType() FileType {
	f := UnknownFileType
	mb, ok := m[MajorBrandKey].(string)
	if ok {
		if mb == "M4A " {
			f = M4A
		} else if mb == "M4B " {
			f = M4B
		} else if mb == "M4P " {
			f = M4P
		}
	}
	return f
}
