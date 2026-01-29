package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/go-drift/drift/pkg/core"
	"github.com/go-drift/drift/pkg/layout"
	"github.com/go-drift/drift/pkg/theme"
	"github.com/go-drift/drift/pkg/widgets"
)

var (
	goLogoOnce  sync.Once
	goLogoImage image.Image
)

func loadImageAsset(name string) image.Image {
	file, err := assetFS.Open("assets/" + name)
	if err != nil {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}
	return img
}

func loadGoLogo() image.Image {
	goLogoOnce.Do(func() {
		goLogoImage = loadImageAsset("go-logo.png")
	})
	return goLogoImage
}

func buildImagesPage(ctx core.BuildContext) core.Widget {
	_, colors, textTheme := theme.UseTheme(ctx)
	logo := loadGoLogo()

	return demoPage(ctx, "Images",
		widgets.Text{Content: "Images", Style: textTheme.TitleLarge},
		widgets.VSpace(12),
		widgets.Text{Content: "Raster images are decoded with Go's image package.", Style: labelStyle(colors)},
		widgets.VSpace(24),
		widgets.RowOf(
			widgets.MainAxisAlignmentCenter,
			widgets.CrossAxisAlignmentStart,
			widgets.MainAxisSizeMax,
			widgets.Image{
				Source: logo,
				Width:  220,
			},
		),
		widgets.VSpace(12),
		widgets.Text{Content: "Go logo (PNG)", Style: textTheme.BodySmall},
		widgets.VSpace(32),
		widgets.Text{Content: "Fit modes", Style: textTheme.TitleMedium},
		widgets.VSpace(12),
		fitPreview("Fill", widgets.ImageFitFill, logo, colors, textTheme),
		widgets.VSpace(16),
		fitPreview("Contain", widgets.ImageFitContain, logo, colors, textTheme),
		widgets.VSpace(16),
		fitPreview("Cover", widgets.ImageFitCover, logo, colors, textTheme),
		widgets.VSpace(16),
		fitPreview("None", widgets.ImageFitNone, logo, colors, textTheme),
		widgets.VSpace(16),
		fitPreview("ScaleDown", widgets.ImageFitScaleDown, logo, colors, textTheme),
		widgets.VSpace(40),
	)
}

func fitPreview(label string, fit widgets.ImageFit, logo image.Image, colors theme.ColorScheme, textTheme theme.TextTheme) core.Widget {
	return widgets.ColumnOf(
		widgets.MainAxisAlignmentStart,
		widgets.CrossAxisAlignmentStart,
		widgets.MainAxisSizeMin,
		widgets.Text{Content: label, Style: textTheme.BodyMedium},
		widgets.VSpace(8),
		widgets.Container{
			Color:     colors.SurfaceVariant,
			Width:     240,
			Height:    140,
			Alignment: layout.AlignmentCenter,
			ChildWidget: widgets.Image{
				Source:    logo,
				Width:     220,
				Height:    120,
				Fit:       fit,
				Alignment: layout.AlignmentCenter,
			},
		},
	)
}
