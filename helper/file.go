package helper

import (
	"bytes"
	"image"
	"io"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-resty/resty/v2"
)

func GetImageFromURL(req *resty.Request, url string) (image.Image, string, error) {
	resp, err := req.Get(url)
	if err != nil {
		return nil, "", err
	}

	reader := bytes.NewReader(resp.Body())

	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func CreateFileFromURL(url string, filepath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func RemoveFile(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return err
	}
	return nil
}

func GetMimeType(reader io.Reader) (buf bytes.Buffer, contentType string, extension string, err error) {
	if _, err = io.Copy(&buf, reader); err != nil {
		return
	}

	mime := mimetype.Detect(buf.Bytes())
	contentType = mime.String()
	extension = mime.Extension()
	return
}
