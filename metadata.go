package yurit

import "time"

// Format is an enumeration of metadata types supported by this package.
type Format string

// Supported tag formats.
const (
	UnknownFormat Format = ""        // Unknown Format.
	ID3v1         Format = "ID3v1"   // ID3v1 tag format.
	ID3v2_2       Format = "ID3v2.2" // ID3v2.2 tag format.
	ID3v2_3       Format = "ID3v2.3" // ID3v2.3 tag format (most common).
	ID3v2_4       Format = "ID3v2.4" // ID3v2.4 tag format.
	MP4           Format = "MP4"     // MP4 tag (atom) format (see http://www.ftyps.com/ for a full file type list)
	VORBIS        Format = "VORBIS"  // Vorbis Comment tag format.
)

// FileType is an enumeration of the audio file types supported by this package, in particular
// there are audio file types which share metadata formats, and this type is used to distinguish
// between them.
type FileType string

// Supported file types.
const (
	UnknownFileType FileType = ""     // Unknown FileType.
	MP1             FileType = "MP1"  // MP1 file
	MP2             FileType = "MP2"  // MP2 file
	MP3             FileType = "MP3"  // MP3 file
	M4A             FileType = "M4A"  // M4A file Apple iTunes (ACC) Audio
	M4B             FileType = "M4B"  // M4A file Apple iTunes (ACC) Audio Book
	M4P             FileType = "M4P"  // M4A file Apple iTunes (ACC) AES Protected Audio
	ALAC            FileType = "ALAC" // Apple Lossless file FIXME: actually detect this
	FLAC            FileType = "FLAC" // FLAC file
	OGG             FileType = "OGG"  // OGG file
	DSF             FileType = "DSF"  // DSF file DSD Sony format see https://dsd-guide.com/sites/default/files/white-papers/DSFFileFormatSpec_E.pdf
)

// Metadata is an interface which is used to describe metadata retrieved by this package.
type Metadata interface {
	// Album returns the album name of the track.
	Album() string
	// AlbumArtist returns the album artist name of the track.
	AlbumArtist() string
	// Artist returns the artist name of the track.
	Artist() string
	// AverageBitrate returns the average bitrate of the file in bits per second
	AverageBitrate() int
	// Comment returns the comment, or an empty string if unavailable.
	Comment() string
	// Composer returns the composer of the track.
	Composer() string
	// Disc returns the disc number and total discs, or zero values if unavailable.
	Disc() (int, int)
	// Duration returns the length of the track as time.
	Duration() time.Duration
	// FileType returns the file type of the audio file.
	FileType() FileType
	// Format returns the metadata Format used to encode the data.
	Format() Format
	// Genre returns the genre of the track.
	Genre() string
	// Lyrics returns the lyrics, or an empty string if unavailable.
	Lyrics() string
	// Picture returns a picture, or nil if not available.
	Picture() *Picture
	// Raw returns the raw mapping of retrieved tag names and associated values.
	// NB: tag/atom names are not standardised between formats.
	Raw() map[string]interface{}
	// Title returns the title of the track.
	Title() string
	// Track returns the track number and total tracks, or zero values if unavailable.
	Track() (int, int)
	// Year returns the year of the track.
	Year() int
}
