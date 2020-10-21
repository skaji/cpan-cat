package cpan

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

var BaseDir = filepath.Join(os.Getenv("HOME"), ".perl-cpm", "sources", "https%cpan.metacpan.org")

type File struct {
	URL   string
	Name  string
	Local string
}

func NewFile(url string) *File {
	name := path.Base(url)
	local := filepath.Join(BaseDir, name)
	return &File{URL: url, Name: name, Local: local}
}

func (f *File) ModTime() time.Time {
	if info, err := os.Stat(f.Local); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

func (f *File) Fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return err
	}
	req.Close = true
	if t := f.ModTime(); !t.IsZero() {
		req.Header.Set("If-Modified-Since", t.Format(http.TimeFormat))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()
	switch res.StatusCode {
	case http.StatusOK:
		lastModified, err := http.ParseTime(res.Header.Get("Last-Modified"))
		if err != nil {
			return err
		}
		tempfile, err := ioutil.TempFile(BaseDir, f.Name)
		if err != nil {
			return err
		}
		defer func() {
			tempfile.Close()
			os.Remove(tempfile.Name())
		}()
		if _, err := io.Copy(tempfile, res.Body); err != nil {
			return err
		}
		if err := tempfile.Chmod(0644); err != nil {
			return err
		}
		if err := tempfile.Close(); err != nil {
			return err
		}
		if err := os.Chtimes(tempfile.Name(), lastModified, lastModified); err != nil {
			return err
		}
		if err := os.Rename(tempfile.Name(), f.Local); err != nil {
			return err
		}
		return nil
	case http.StatusNotModified:
		return nil
	default:
		return errors.New(res.Status)
	}
}

func (f *File) Cat(ctx context.Context, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "gzip", "-dc", f.Local)
	cmd.Stdout = w
	return cmd.Run()
}
