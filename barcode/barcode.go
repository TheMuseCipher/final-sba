package barcode

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func DecodeBarcodeFromImage(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Convert image to binary bitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("failed to create binary bitmap: %v", err)
	}

	// Try QR code first
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try Code128
	reader := oned.NewCode128Reader()
	result, err = reader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try EAN13
	eanReader := oned.NewEAN13Reader()
	result, err = eanReader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try EAN8
	ean8Reader := oned.NewEAN8Reader()
	result, err = ean8Reader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	// Try UPC-A
	upcReader := oned.NewUPCAReader()
	result, err = upcReader.Decode(bmp, nil)
	if err == nil && result != nil {
		return result.String(), nil
	}

	return "", fmt.Errorf("could not decode barcode or QR code from image")
}

