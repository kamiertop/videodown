package main

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"image/png"
	"os"
	"testing"
)

func TestWindowsIconResourcesContain256PixelPNG(t *testing.T) {
	source, err := os.ReadFile("../../internal/gui/assets/appicon.png")
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"amd64", "arm64", "386"} {
		t.Run(arch, func(t *testing.T) {
			object, err := windowsIconObject(source, arch)
			if err != nil {
				t.Fatal(err)
			}
			file, err := pe.NewFile(bytes.NewReader(object))
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			section := file.Section(".rsrc")
			if section == nil || len(section.Relocs) != 2 {
				t.Fatal("missing icon resource relocations")
			}
			raw, err := section.Data()
			if err != nil {
				t.Fatal(err)
			}
			le := binary.LittleEndian
			if le.Uint16(raw[14:]) != 2 || le.Uint32(raw[16:]) != 3 || le.Uint32(raw[24:]) != 14 {
				t.Fatal("missing RT_ICON or RT_GROUP_ICON")
			}
			at, size := le.Uint32(raw[128:]), le.Uint32(raw[132:])
			icon, err := png.DecodeConfig(bytes.NewReader(raw[at : at+size]))
			if err != nil || icon.Width != 256 || icon.Height != 256 {
				t.Fatalf("invalid Windows PNG icon: %+v %v", icon, err)
			}
		})
	}
}
