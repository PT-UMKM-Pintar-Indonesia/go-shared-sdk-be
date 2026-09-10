package sdk_helper

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

const (
	maxQRBase64Length = 2 * 1024 * 1024
	maxQRImageWidth   = 2000
	maxQRImageHeight  = 2000
	minQRImageSize    = 50
)

func EmvcoQris(qris string) (string, error) {
	if len(qris) == 0 {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: empty input")
	}

	if len(qris) > maxQRBase64Length {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: input too large (%d bytes)", len(qris))
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
	w, h := bounds.Dx(), bounds.Dy()

	if w < minQRImageSize || h < minQRImageSize {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: image too small (%dx%d), not a valid QR", w, h)
	}

	if w > maxQRImageWidth || h > maxQRImageHeight {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: image too large (%dx%d)", w, h)
	}

	qrImageBitmap, err := gozxing.NewBinaryBitmapFromImage(qrImage)
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to create bitmap: %w", err)
	}

	hints := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true}

	result, err := qrcode.NewQRCodeReader().Decode(qrImageBitmap, hints)
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: QR code unreadable (%dx%d, %s): %w", w, h, format, err)
	}

	return result.GetText(), nil
}
