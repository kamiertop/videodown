package imagecache

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/netclient"
)

func orientedJPEG(t *testing.T, orientation uint16, order binary.ByteOrder) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 800, 400)), nil); err != nil {
		t.Fatal(err)
	}
	tiff := make([]byte, 26)
	copy(tiff, "II")
	if order == binary.BigEndian {
		copy(tiff, "MM")
	}
	order.PutUint16(tiff[2:], 42)
	order.PutUint32(tiff[4:], 8)
	order.PutUint16(tiff[8:], 1)
	order.PutUint16(tiff[10:], 0x112)
	order.PutUint16(tiff[12:], 3)
	order.PutUint32(tiff[14:], 1)
	order.PutUint16(tiff[18:], orientation)
	segment := append([]byte("Exif\x00\x00"), tiff...)
	result := []byte{0xff, 0xd8, 0xff, 0xe1, 0, byte(len(segment) + 2)}
	result = append(result, segment...)
	return append(result, buf.Bytes()[2:]...)
}

func TestThumbnailOrientation(t *testing.T) {
	src := image.NewRGBA(image.Rect(10, 20, 13, 22))
	for i := range 6 {
		src.SetRGBA(10+i%3, 20+i/3, color.RGBA{R: uint8(i + 1), A: 255})
	}
	expected := [][]uint8{{1, 2, 3, 4, 5, 6}, {3, 2, 1, 6, 5, 4}, {6, 5, 4, 3, 2, 1}, {4, 5, 6, 1, 2, 3}, {1, 4, 2, 5, 3, 6}, {4, 1, 5, 2, 6, 3}, {6, 3, 5, 2, 4, 1}, {3, 6, 2, 5, 1, 4}}
	for o := 1; o <= 8; o++ {
		for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
			if got := jpegOrientation(orientedJPEG(t, uint16(o), order)); got != o {
				t.Fatalf("orientation: got %d want %d", got, o)
			}
		}
		got := orientThumbnail(src, o)
		var pixels []uint8
		for y := got.Bounds().Min.Y; y < got.Bounds().Max.Y; y++ {
			for x := got.Bounds().Min.X; x < got.Bounds().Max.X; x++ {
				pixels = append(pixels, color.RGBAModel.Convert(got.At(x, y)).(color.RGBA).R)
			}
		}
		if !reflect.DeepEqual(pixels, expected[o-1]) {
			t.Fatalf("orientation %d: got %v", o, pixels)
		}
	}
}

func TestLoadThumbnailKeepsDirectionAndSizeBound(t *testing.T) {
	data := orientedJPEG(t, 6, binary.LittleEndian)
	previous := imageClient
	imageClient = netclient.New("", 0, imageRoundTripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	}))
	defer func() { imageClient = previous }()
	thumbnail, err := LoadThumbnail(context.Background(), "https://fixture.test/cover.jpg", "")
	if err != nil {
		t.Fatal(err)
	}
	if w, h := thumbnail.Size(); w != 256 || h != 512 {
		t.Fatalf("thumbnail: %dx%d", w, h)
	}
	original, err := Load(context.Background(), "https://fixture.test/cover.jpg", "")
	if err != nil {
		t.Fatal(err)
	}
	if w, h := original.Size(); w != 400 || h != 800 {
		t.Fatalf("original: %dx%d", w, h)
	}
	for i := range len(data) {
		jpegOrientation(data[:i])
	} // Truncated metadata must not panic.
}

func TestNearViewportPreloadsOnlyNearbyRows(t *testing.T) {
	for _, tc := range []struct {
		y    float32
		want bool
	}{{100, true}, {900, true}, {-200, true}, {1400, false}, {-800, false}} {
		if got := NearViewport(ui.Rect{X: 10, Y: tc.y, W: 200, H: 150}, 1200, 800); got != tc.want {
			t.Fatalf("y=%v: %v", tc.y, got)
		}
	}
	if NearViewport(ui.Rect{}, 1200, 800) {
		t.Fatal("unlaid-out images must not load")
	}
}

type imageRoundTripper func(*http.Request) (*http.Response, error)

func (f imageRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
