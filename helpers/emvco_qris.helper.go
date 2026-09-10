package sdk_helper

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func EmvcoQris(qris string) (string, error) {
	if len(qris) == 0 {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: empty input")
	}

	base64Str := qris
	if idx := strings.Index(qris, ","); idx != -1 {
		base64Str = qris[idx+1:]
	}

	base64Str = strings.TrimSpace(strings.ReplaceAll(base64Str, "\n", ""))
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

	qrImage, format, err := image.Decode(bytes.NewReader(decodedBytes))
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to decode image (format=%s): %w", format, err)
	}

	bounds := qrImage.Bounds()
	padding := bounds.Dx() / 10
	if padding < 10 {
		padding = 10
	}

	newWidth := bounds.Dx() + (padding * 2)
	newHeight := bounds.Dy() + (padding * 2)

	rgba := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(rgba, image.Rect(padding, padding, newWidth-padding, newHeight-padding), qrImage, bounds.Min, draw.Src)

	hints := map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER:       true,
		gozxing.DecodeHintType_CHARACTER_SET:    "UTF-8",
		gozxing.DecodeHintType_POSSIBLE_FORMATS: []gozxing.BarcodeFormat{gozxing.BarcodeFormat_QR_CODE},
	}

	reader := qrcode.NewQRCodeReader()

	bitmap1, err := gozxing.NewBinaryBitmapFromImage(rgba)
	if err == nil {
		if result, err := reader.Decode(bitmap1, hints); err == nil {
			return result.GetText(), nil
		}
	}

	source := gozxing.NewLuminanceSourceFromImage(rgba)
	bitmap2, err := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(source))
	if err == nil {
		if result, err := reader.Decode(bitmap2, hints); err == nil {
			return result.GetText(), nil
		}
	}

	return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: QR code genuinely unreadable after preprocessing and dual binarizer attempts")
}
