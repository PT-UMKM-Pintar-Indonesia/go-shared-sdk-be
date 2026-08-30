package sdk_helper

import (
	"bytes"
	"encoding/base64"
	"image"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func EmvcoQris(qris string) (string, error) {
	base64Str := qris

	if idx := bytes.Index([]byte(qris), []byte(",")); idx != -1 {
		base64Str = qris[idx+1:]
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	qrImage, _, err := image.Decode(bytes.NewReader(decodedBytes))
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	qrImageBitmap, err := gozxing.NewBinaryBitmapFromImage(qrImage)
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	qrImageDecode, err := qrcode.NewQRCodeReader().Decode(qrImageBitmap, nil)
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	return qrImageDecode.GetText(), nil
}
