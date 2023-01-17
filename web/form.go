package web

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
)

func MultipartTestForm(filefield string, filename string, extrafields map[string]string) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	var fw io.Writer

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return b, nil, err
		}
		if fw, err = w.CreateFormFile(filefield, file.Name()); err != nil {
			return b, nil, err
		}
		if _, err = io.Copy(fw, file); err != nil {
			return b, nil, err
		}
	}

	for k, v := range extrafields {
		f, err := w.CreateFormField(k)
		if err != nil {
			return b, nil, err
		}
		if _, err := io.Copy(f, bytes.NewBufferString(v)); err != nil {
			return b, nil, err
		}
	}

	if err := w.Close(); err != nil {
		return b, nil, err
	}
	return b, w, nil
}
