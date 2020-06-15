package yurit

import (
	"bytes"
	"fmt"
	"strconv"
)

type mp4metadata map[string]interface{}

func processILSTAtom(ilst Mp4Atom) (mp4metadata, error) {
	m := mp4metadata{}
	for _, metadataItem := range ilst.Children {
		name := metadataItem.Name
		if name == "----" {
			mean := ""
			meanAtom := findAtom(metadataItem.Children, "mean")
			if meanAtom != nil && len(meanAtom.Data) > 4 {
				//Skip version/flag bytes, and just get mean data as string
				mean = string(meanAtom.Data[4:])
			}
			nameAtom := findAtom(metadataItem.Children, "name")
			if nameAtom != nil && len(nameAtom.Data) > 4 {
				//Skip version/flag bytes, and just get name data as string
				name = string(nameAtom.Data[4:])
			}
			//Verify mean is expected value and name has changed
			if mean != "com.apple.iTunes" || name == "----" {
				continue
			}
		}
		//Find first data atom
		dataAtom := findAtom(metadataItem.Children, "data")
		if len(dataAtom.Data) < 8 {
			return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for atom version and flags, got %d", 8, len(dataAtom.Data))
		}
		dataType := getUint24AsInt(dataAtom.Data[1:4])
		dataPortion := dataAtom.Data[8:]
		if name == "trkn" || name == "disk" {
			if len(dataPortion) < 6 {
				return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for track and disk numbers, got %d", 6, len(dataPortion))
			}
			m[name] = int(dataPortion[3])
			m[name+"_count"] = int(dataPortion[5])
		} else if name == "covr" {
			contentType := ""
			if dataType == 13 {
				contentType = "jpeg"
			} else if dataType == 14 || bytes.HasPrefix(dataPortion, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
				contentType = "png"
			}
			if contentType == "" {
				continue
			}
			picture := &Picture{
				Ext:      contentType,
				MIMEType: "image/" + contentType,
				Data:     dataPortion,
			}
			m[name] = picture
		} else if dataType == 0 {
			m[name] = dataPortion
		} else if dataType == 1 {
			m[name] = string(dataPortion)
		} else if dataType == 21 {
			if len(dataPortion) < 1 {
				return nil, fmt.Errorf("invalid encoding: expected at least %d bytes for integer tag data, got %d", 1, len(dataPortion))
			}
			m[name] = int(dataPortion[0])
		}
	}
	return m, nil
}

func (m mp4metadata) Album() string {
	t, _ := m["\xa9alb"].(string)
	return t
}

func (m mp4metadata) AlbumArtist() string {
	t, _ := m["aART"].(string)
	return t
}

func (m mp4metadata) Artist() string {
	t, ok := m["\xa9art"].(string)
	if !ok {
		t, _ = m["\xa9ART"].(string)
	}
	return t
}

func (m mp4metadata) Comment() string {
	t, _ := m["\xa9cmt"].(string)
	return t
}

func (m mp4metadata) Composer() string {
	t, _ := m["\xa9wrt"].(string)
	return t
}

func (m mp4metadata) Disc() (int, int) {
	x, _ := m["disk"].(int)
	y, _ := m["disk_count"].(int)
	return x, y
}

func (m mp4metadata) Genre() string {
	t, _ := m["\xa9gen"].(string)
	return t
}

func (m mp4metadata) Lyrics() string {
	t, _ := m["\xa9lyr"].(string)
	return t
}

func (m mp4metadata) Picture() *Picture {
	p, _ := m["covr"].(*Picture)
	return p
}

func (m mp4metadata) Title() string {
	t, _ := m["\xa9nam"].(string)
	return t
}

func (m mp4metadata) Track() (int, int) {
	x, _ := m["trkn"].(int)
	y, _ := m["trkn_count"].(int)
	return x, y
}

func (m mp4metadata) Year() int {
	var year int = 0
	t, ok := m["\xa9day"].(string)
	if ok && len(t) >= 4 {
		year, _ = strconv.Atoi(t[:4])
	}
	return year
}
