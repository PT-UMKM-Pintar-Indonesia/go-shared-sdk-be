package sdk_helper

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	stdDraw "image/draw"
	"image/png"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func TestEmvcoQrisInputValidation(t *testing.T) {
	if _, err := EmvcoQris(""); err == nil || !strings.Contains(err.Error(), "empty input") {
		t.Fatalf("empty got %v", err)
	}

	// Pass-through for raw EMVCo payloads.
	raw := "00020101021226610014ID.CO.QRIS.WWW0214ID12345678905802ID5907MERCHANT6007CITY0503150"
	got, err := EmvcoQris("  " + raw + "  ")
	if err != nil || got != raw {
		t.Fatalf("passthrough got %q err %v", got, err)
	}

	if _, err := EmvcoQris(strings.Repeat("A", MaxBase64Size+1)); err == nil ||
		!strings.Contains(err.Error(), "exceeds maximum limit") {
		t.Fatalf("oversize got %v", err)
	}

	if _, err := EmvcoQris("data:image/png;base64,"); err == nil ||
		!strings.Contains(err.Error(), "payload empty after cleanup") {
		t.Fatalf("empty payload got %v", err)
	}

	if _, err := EmvcoQris("!!!not base64!!!"); err == nil ||
		!strings.Contains(err.Error(), "failed to decode base64") {
		t.Fatalf("bad base64 got %v", err)
	}

	notImage := base64.StdEncoding.EncodeToString([]byte("just text, not an image"))
	if _, err := EmvcoQris(notImage); err == nil ||
		!strings.Contains(err.Error(), "failed to decode image") {
		t.Fatalf("non-image got %v", err)
	}
}

func TestEmvcoQrisRejectsOversizedImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, MaxDimension+1, MaxDimension+1))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	_, err := EmvcoQris(base64.StdEncoding.EncodeToString(buf.Bytes()))
	if err == nil || !strings.Contains(err.Error(), "dimensions exceed") {
		t.Fatalf("got %v", err)
	}
}

func encodeQRPNG(t *testing.T, payload string, size int) string {
	t.Helper()

	writer := qrcode.NewQRCodeWriter()

	matrix, err := writer.Encode(payload, gozxing.BarcodeFormat_QR_CODE, size, size, nil)
	if err != nil {
		t.Fatalf("encode qr: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	stdDraw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, stdDraw.Src)
	stdDraw.Draw(img, matrix.Bounds(), matrix, image.Point{}, stdDraw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestEmvcoQrisDecodesQRImage(t *testing.T) {
	const payload = "5802ID9907TESTMER6007CITY0503150"

	b64 := encodeQRPNG(t, payload, 450)

	got, err := EmvcoQris(b64)
	if err != nil {
		t.Fatalf("EmvcoQris: %v", err)
	}

	if got != payload {
		t.Fatalf("decoded %q want %q", got, payload)
	}

	// Data-URI style prefix must be stripped at the first comma.
	got, err = EmvcoQris("data:image/png;base64," + b64)
	if err != nil || got != payload {
		t.Fatalf("prefixed decode got %q err %v", got, err)
	}

	// Small images are upscaled before decoding, still readable.
	got, err = EmvcoQris(encodeQRPNG(t, payload, 200))
	if err != nil || got != payload {
		t.Fatalf("upscaled decode got %q err %v", got, err)
	}
}

func TestEmvcoQrisUnreadableImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 420, 420))
	stdDraw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, stdDraw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	_, err := EmvcoQris(base64.StdEncoding.EncodeToString(buf.Bytes()))
	if err == nil || !strings.Contains(err.Error(), "genuinely unreadable") {
		t.Fatalf("got %v", err)
	}
}

func TestEmvcoInternalImageHelpers(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 2))
	src.Set(0, 0, color.RGBA{0, 0, 0, 255})
	src.Set(3, 1, color.RGBA{255, 255, 255, 255})

	padded := addPadding(src, 5)
	if got := padded.Bounds().Dx(); got != 14 {
		t.Fatalf("padded width %d", got)
	}

	if got := padded.Bounds().Dy(); got != 12 {
		t.Fatalf("padded height %d", got)
	}

	if !isWhite(padded.At(0, 0)) {
		t.Fatal("padding must be white")
	}

	if !isBlack(padded.At(5, 5)) {
		t.Fatal("top-left origin pixel must stay black at pad offset")
	}

	rotated := rotate90(src, 1)
	if rotated.Bounds().Dx() != 2 || rotated.Bounds().Dy() != 4 {
		t.Fatalf("rotation swapped dims wrong: %v", rotated.Bounds())
	}

	// Four rotations return the original pixel layout.
	identity := rotate90(src, 4)
	if identity.Bounds() != src.Bounds() {
		t.Fatalf("4x rotate90 changed bounds")
	}

	if isWhite(identity.At(0, 0)) {
		t.Fatal("4x rotate90 must preserve pixel placement")
	}

	if got := rotate90(src, 0); got != src {
		t.Fatal("zero rotation must return src")
	}

	enhanced := enhanceContrast(src)
	if enhanced.Bounds() != src.Bounds() {
		t.Fatal("contrast keeps bounds")
	}

	if !isBlack(enhanced.At(0, 0)) || !isWhite(enhanced.At(3, 1)) {
		t.Fatal("contrast must binarize pixels")
	}
}

func isWhite(c color.Color) bool {
	r, g, b, _ := c.RGBA()

	return r > 0x8000 && g > 0x8000 && b > 0x8000
}

func isBlack(c color.Color) bool {
	r, g, b, _ := c.RGBA()

	return r < 0x8000 && g < 0x8000 && b < 0x8000
}
