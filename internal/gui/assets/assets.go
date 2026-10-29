// Package assets owns the native UI's embedded images and icons.
package assets

import (
	_ "embed"
	"github.com/egoist/mygo/ui"
)

const AppName = "videodown"
const RepositoryURL = "https://github.com/kamiertop/videodown"

//go:embed appicon.png
var AppIconPNG []byte

//go:embed bilibili.svg
var bilibiliSVG []byte

//go:embed douyin.svg
var douyinSVG []byte

//go:embed github.svg
var githubSVG []byte
var GitHub = ui.MustParseSVG(githubSVG)
var platforms = map[string]*ui.SVG{"bilibili": ui.MustParseSVG(bilibiliSVG), "douyin": ui.MustParseSVG(douyinSVG)}
var AppIcon = func() *ui.Bitmap {
	image, err := ui.DecodeBitmap(AppIconPNG)
	if err != nil {
		panic(err)
	}
	return image
}()

func Platform(id string) *ui.SVG { return platforms[id] }

// Native photo and Live Photo badge paths.
//
//go:embed douyin-live-photo.svg
var douyinLivePhotoSVG []byte
var DouyinLivePhoto = ui.MustParseSVG(douyinLivePhotoSVG)

//go:embed douyin-image.svg
var douyinImageSVG []byte
var DouyinImage = ui.MustParseSVG(douyinImageSVG)
