package imagecache

import (
	"bytes"
	"encoding/binary"
	"image"
)

// jpegOrientation reads only the first TIFF IFD, with all offsets bounded by
// the APP1 segment. Invalid or absent metadata leaves the image unchanged.
func jpegOrientation(data []byte) int {
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}
	for i := 2; i+4 <= len(data); {
		if data[i] != 0xff {
			return 1
		}
		marker := data[i+1]
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		size := int(binary.BigEndian.Uint16(data[i+2:]))
		if size < 2 || size > len(data)-i-2 {
			return 1
		}
		segment := data[i+4 : i+2+size]
		if marker == 0xe1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return tiffOrientation(segment[6:])
		}
		i += size + 2
	}
	return 1
}

func tiffOrientation(data []byte) int {
	if len(data) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	if order.Uint16(data[2:]) != 42 {
		return 1
	}
	offset := uint64(order.Uint32(data[4:]))
	if offset < 8 || offset+2 > uint64(len(data)) {
		return 1
	}
	count := int(order.Uint16(data[offset:]))
	for pos := int(offset) + 2; count > 0 && pos <= len(data)-12; pos, count = pos+12, count-1 {
		entry := data[pos : pos+12]
		if order.Uint16(entry) != 0x112 {
			continue
		}
		if order.Uint16(entry[2:]) != 3 || order.Uint32(entry[4:]) != 1 {
			return 1
		}
		value := int(order.Uint16(entry[8:]))
		if value >= 1 && value <= 8 {
			return value
		}
		return 1
	}
	return 1
}

// Rotate after downsampling so orientation never allocates a full-size copy.
func orientThumbnail(src image.Image, orientation int) image.Image {
	if orientation < 2 || orientation > 8 {
		return src
	}
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := range dh {
		for x := range dw {
			sx, sy := x, y
			switch orientation {
			case 2:
				sx, sy = w-1-x, y
			case 3:
				sx, sy = w-1-x, h-1-y
			case 4:
				sx, sy = x, h-1-y
			case 5:
				sx, sy = y, x
			case 6:
				sx, sy = y, h-1-x
			case 7:
				sx, sy = w-1-y, h-1-x
			case 8:
				sx, sy = w-1-y, x
			}
			dst.Set(x, y, src.At(bounds.Min.X+sx, bounds.Min.Y+sy))
		}
	}
	return dst
}
