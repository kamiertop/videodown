package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"

	"golang.org/x/image/draw"
)

// windowsIconObject creates a COFF .rsrc object using the standard Windows
// RT_ICON / RT_GROUP_ICON resource tree. Go links it without a C toolchain.
func windowsIconObject(source []byte, arch string) ([]byte, error) {
	machine, relocation := uint16(0), uint16(0)
	switch arch {
	case "amd64":
		machine, relocation = 0x8664, 3
	case "arm64":
		machine, relocation = 0xaa64, 2
	case "386":
		machine, relocation = 0x14c, 7
	default:
		return nil, fmt.Errorf("不支持的 Windows 图标架构: %s", arch)
	}
	original, err := png.Decode(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}
	scaled := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), original, original.Bounds(), draw.Src, nil)
	var icon bytes.Buffer
	if err := png.Encode(&icon, scaled); err != nil {
		return nil, err
	}
	le := binary.LittleEndian
	// An icon group header followed by one 256x256, 32-bit PNG icon entry.
	group := make([]byte, 20)
	le.PutUint16(group[2:], 1)
	le.PutUint16(group[4:], 1)
	le.PutUint16(group[10:], 1)
	le.PutUint16(group[12:], 32)
	le.PutUint32(group[14:], uint32(icon.Len()))
	le.PutUint16(group[18:], 1)
	payloads := [][]byte{icon.Bytes(), group}
	// Two type directories, each containing resource ID 1 and language 0x0409.
	// Directory headers are 16 bytes; directory entries are 8; data entries are 16.
	const rootSize = 32
	const typesAt = rootSize
	const languagesAt = typesAt + 2*24
	const entriesAt = languagesAt + 2*24
	const payloadAt = entriesAt + 2*16
	raw := make([]byte, payloadAt)
	le.PutUint16(raw[14:], 2)
	for i, kind := range []uint32{3, 14} {
		typeAt, languageAt, entryAt := typesAt+i*24, languagesAt+i*24, entriesAt+i*16
		le.PutUint32(raw[16+i*8:], kind)
		le.PutUint32(raw[20+i*8:], uint32(typeAt)|0x80000000)
		le.PutUint16(raw[typeAt+14:], 1)
		le.PutUint32(raw[typeAt+16:], 1)
		le.PutUint32(raw[typeAt+20:], uint32(languageAt)|0x80000000)
		le.PutUint16(raw[languageAt+14:], 1)
		le.PutUint32(raw[languageAt+16:], 0x0409)
		le.PutUint32(raw[languageAt+20:], uint32(entryAt))
		for len(raw)%4 != 0 {
			raw = append(raw, 0)
		}
		le.PutUint32(raw[entryAt:], uint32(len(raw)))
		le.PutUint32(raw[entryAt+4:], uint32(len(payloads[i])))
		raw = append(raw, payloads[i]...)
	}
	for len(raw)%4 != 0 {
		raw = append(raw, 0)
	}
	const rawAt = 60 // 20-byte COFF header followed by one 40-byte section header.
	relocAt := rawAt + len(raw)
	symbolsAt := relocAt + 2*10
	object := make([]byte, symbolsAt+18+4)
	le.PutUint16(object, machine)
	le.PutUint16(object[2:], 1)
	le.PutUint32(object[8:], uint32(symbolsAt))
	le.PutUint32(object[12:], 1)
	copy(object[20:], ".rsrc")
	le.PutUint32(object[36:], uint32(len(raw)))
	le.PutUint32(object[40:], rawAt)
	le.PutUint32(object[44:], uint32(relocAt))
	le.PutUint16(object[52:], 2)
	le.PutUint32(object[56:], 0x40000040) // initialized, read-only section data
	copy(object[rawAt:], raw)
	for i := range 2 {
		at := relocAt + i*10
		le.PutUint32(object[at:], uint32(entriesAt+i*16))
		le.PutUint16(object[at+8:], relocation)
	}
	copy(object[symbolsAt:], ".rsrc")
	le.PutUint16(object[symbolsAt+12:], 1)
	object[symbolsAt+16] = 3
	le.PutUint32(object[symbolsAt+18:], 4) // empty COFF string table
	return object, nil
}
