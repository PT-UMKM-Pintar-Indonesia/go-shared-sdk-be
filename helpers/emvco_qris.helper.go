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

	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER:    true,
		gozxing.DecodeHintType_CHARACTER_SET: "UTF-8",
	}

	reader := qrcode.NewQRCodeReader()

	bitmap1, err := gozxing.NewBinaryBitmapFromImage(qrImage)
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to create bitmap: %w", err)
	}

	result, err := reader.Decode(bitmap1, hints)
	if err == nil {
		return result.GetText(), nil
	}

	source := gozxing.NewLuminanceSourceFromImage(qrImage)

	bitmap2, err := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(source))
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: failed to create fallback bitmap: %w", err)
	}

	result, err = reader.Decode(bitmap2, hints)
	if err == nil {
		return result.GetText(), nil
	}

	bounds := qrImage.Bounds()

	return sdk_cons.EMPTY, fmt.Errorf("EmvcoQris: QR code unreadable (%dx%d, %s)", bounds.Dx(), bounds.Dy(), format)
}
