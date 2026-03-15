package audit

const (
	AELMagic     = "AEL0"
	AELVersion   = uint16(0)
	AELHeaderLen = uint16(64)
)

// SegmentHeader is the fixed 64-byte binary header for each audit log segment.
// Layout: 4+2+2+4+8+8+32+4 = 64 bytes.
// Use encoding/binary with binary.LittleEndian when reading/writing to disk.
type SegmentHeader struct {
	Magic           [4]byte
	Version         uint16
	HeaderLen       uint16
	Flags           uint32
	CreatedAtUnixMs int64
	SegmentID       uint64
	Reserved        [32]byte // zero-filled
	HeaderCRC32     uint32
}
