package imgconv

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Converter struct {
	Directory string
	FromExt   string
	ToExt     string
}

var supported = map[string]struct{}{
	"jpg":  struct{}{},
	"jpeg": struct{}{},
	"png":  struct{}{},
	"gif":  struct{}{},
}

func New(directory, from, to string) (*Converter, error) {
	if err := validate(directory, strings.ToLower(from), strings.ToLower(to)); err != nil {
		return nil, err
	}
	return &Converter{directory, from, to}, nil
}

func validate(directory, from, to string) error {
	src := filepath.Clean(directory)
	info, err := os.Stat(src)

	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("%s does not exist: %w", src, err)
			return err
		}
		return fmt.Errorf("%s failed to get information: %w", src, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s must be directory", src)
	}

	if _, ok := supported[from]; !ok {
		return fmt.Errorf("unsupported format %s", from)
	}

	if normalizedExt(from) == normalizedExt(to) {
		return fmt.Errorf("%s should be different from %s", from, to)
	}
	return nil
}

func (c *Converter) Walk() error {
	return filepath.Walk(c.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := normalizedExt(path)
		if ext != c.FromExt {
			return nil
		}

		reader, rErr := os.Open(path)
		if rErr != nil {
			return fmt.Errorf("failed to open %s: %v", path, rErr)
		}

		defer reader.Close()
		dest := strings.TrimSuffix(path, filepath.Ext(path)) + "." + c.ToExt
		writer, wErr := os.Create(dest)
		if wErr != nil {
			return fmt.Errorf("failed to write to %s: %v", dest, wErr)
		}
		defer writer.Close()
		return c.convert(writer, reader)
	})
}

func (c *Converter) convert(w io.Writer, r io.Reader) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	switch c.ToExt {
	case "jpg", "jpeg":
		return jpeg.Encode(w, img, nil)

	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	default:
		return fmt.Errorf("unkown format %s", c.ToExt)

	}
}

func normalizedExt(path_or_ext string) string {
	ext := strings.TrimLeft(filepath.Ext(path_or_ext), ".")
	if ext == "" {
		ext = path_or_ext
	}
	ext = strings.ToLower(ext)
	switch ext {
	case "jpeg", "jpg":
		return "jpg"
	default:
		return ext
	}
}
