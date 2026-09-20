package sdk_helper

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	stdDraw "image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
)

const (
	MaxBase64Size = 10 * 1024 * 1024
	MaxDimension  = 4090
	MinDimension  = 400
)

func addPadding(src *image.RGBA, pad int) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w+(pad*2), h+(pad*2)))

	stdDraw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, stdDraw.Src)
	stdDraw.Draw(dst, image.Rect(pad, pad, w+pad, h+pad), src, bounds.Min, stdDraw.Over)

	return dst
}

func rotate90(src *image.RGBA, times int) *image.RGBA {
	times = (times%4 + 4) % 4
	if times == 0 {
		return src
	}

	current := src
	for i := 0; i < times; i++ {
		bounds := current.Bounds()
		w, h := bounds.Dx(), bounds.Dy()
		dst := image.NewRGBA(image.Rect(0, 0, h, w))

		srcPix := current.Pix
		dstPix := dst.Pix
		srcStride := current.Stride
		dstStride := dst.Stride

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				srcOffset := y*srcStride + x*4
				dstOffset := x*dstStride + (h-1-y)*4

				copy(dstPix[dstOffset:dstOffset+4], srcPix[srcOffset:srcOffset+4])
			}
		}
		current = dst
	}

	return current
}

func enhanceContrast(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	srcPix := src.Pix
	dstPix := dst.Pix

	for i := 0; i < len(srcPix); i += 4 {
		r := uint32(srcPix[i])
		g := uint32(srcPix[i+1])
		b := uint32(srcPix[i+2])
		a := srcPix[i+3]

		lum := (299*r + 587*g + 114*b) / 1000

		var val uint8 = 255
		if lum < 128 {
			val = 0
		}

		dstPix[i] = val
		dstPix[i+1] = val
		dstPix[i+2] = val
		dstPix[i+3] = a
	}

	return dst
}

func EmvcoQris(qris string) (string, error) {
	if len(qris) == 0 {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: empty input")
	}

	cleanInput := strings.TrimSpace(qris)

	if strings.HasPrefix(cleanInput, "000201") {
		return cleanInput, nil
	}

	if len(cleanInput) > MaxBase64Size {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: payload exceeds maximum limit of 10MB")
	}

	base64Str := cleanInput
	if idx := strings.Index(cleanInput, ","); idx != -1 {
		base64Str = cleanInput[idx+1:]
	}

	base64Str = strings.TrimSpace(strings.ReplaceAll(base64Str, "\n", ""))
	base64Str = strings.ReplaceAll(base64Str, "\r", "")
	base64Str = strings.ReplaceAll(base64Str, "-", "+")
	base64Str = strings.ReplaceAll(base64Str, "_", "/")

	if len(base64Str) == 0 {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: base64 payload empty after cleanup")
	}

	if m := len(base64Str) % 4; m != 0 {
		base64Str += strings.Repeat("=", 4-m)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to decode base64: %w", err)
	}

	srcImg, format, err := image.Decode(bytes.NewReader(decodedBytes))
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to decode image (format=%s): %w", format, err)
	}

	bounds := srcImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width > MaxDimension || height > MaxDimension {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: image dimensions exceed maximum allowed (%dx%d)", width, height)
	}

	rgbaImg := image.NewRGBA(image.Rect(0, 0, width, height))
	stdDraw.Draw(rgbaImg, rgbaImg.Bounds(), &image.Uniform{color.White}, image.Point{}, stdDraw.Src)
	stdDraw.Draw(rgbaImg, rgbaImg.Bounds(), srcImg, bounds.Min, stdDraw.Over)

	if width < MinDimension || height < MinDimension {
		scale := float64(MinDimension) / float64(width)
		if float64(MinDimension)/float64(height) > scale {
			scale = float64(MinDimension) / float64(height)
		}

		newW := int(float64(width) * scale)
		newH := int(float64(height) * scale)

		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		stdDraw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, stdDraw.Src)

		xdraw.BiLinear.Scale(dst, dst.Bounds(), rgbaImg, rgbaImg.Bounds(), stdDraw.Over, nil)
		rgbaImg = dst
	}

	hints := map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER:       true,
		gozxing.DecodeHintType_CHARACTER_SET:    "UTF-8",
		gozxing.DecodeHintType_POSSIBLE_FORMATS: []gozxing.BarcodeFormat{gozxing.BarcodeFormat_QR_CODE},
	}

	imgWithPadding := addPadding(rgbaImg, 30)
	reader := qrcode.NewQRCodeReader()

	candidates := []*image.RGBA{
		imgWithPadding,
		enhanceContrast(imgWithPadding),
	}

	for _, imgCandidate := range candidates {
		for rotationPass := 0; rotationPass < 4; rotationPass++ {
			rotated := imgCandidate
			if rotationPass > 0 {
				rotated = rotate90(imgCandidate, rotationPass)
			}

			source := gozxing.NewLuminanceSourceFromImage(rotated)

			if bitmap, err := gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(source)); err == nil {
				if result, err := reader.Decode(bitmap, hints); err == nil {
					return result.GetText(), nil
				}
			}

			if bitmap, err := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(source)); err == nil {
				if result, err := reader.Decode(bitmap, hints); err == nil {
					return result.GetText(), nil
				}
			}

			if bitmap, err := gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(source.Invert())); err == nil {
				if result, err := reader.Decode(bitmap, hints); err == nil {
					return result.GetText(), nil
				}
			}
		}
	}

	return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: QR code genuinely unreadable after preprocessing and multi-strategy decoding")
}
