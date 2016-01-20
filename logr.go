package logr

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// TimeFormat is the default format used for the suffix date and time on each rotated log.
	TimeFormat = "2006-01-02_1504"
)

// RotatingWriter is a io.Writer which wraps a *os.File, suitable for log rotation.
type RotatingWriter struct {
	lock        sync.Mutex
	filename    string
	file        *os.File
	currentSize int64
	startDate   time.Time

	timeFormat string
	prefix     bool
	daily      bool
	compress   bool
	maxSize    int64
}

// NewWriter creates a new file and returns a rotating writer.
func NewWriter(filename string) (*RotatingWriter, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	return NewWriterFromFile(file)
}

// NewWriterWithCompression creates a new file and returns a rotating writer compressing
// the old files.
func NewWriterWithCompression(filename string) (*RotatingWriter, error) {
	w, err := NewWriter(filename)
	w.compress = true
	return w, err
}

// NewWriterFromFile creates a rotating writer using the provided file as base.
//
// The caller must take care to not close the file it provides here, as the RotatingWriter
// will do it automatically when rotating.
func NewWriterFromFile(file *os.File) (*RotatingWriter, error) {
	w := &RotatingWriter{
		filename:  file.Name(),
		file:      file,
		maxSize:   -1,
		startDate: time.Now(),
	}

	if err := w.readCurrentSize(); err != nil {
		return nil, err
	}

	return w, nil
}

// NewWriterFromFileWithCompression is the same as NewWriteFromFile but with
// compression enabled.
func NewWriterFromFileWithCompression(file *os.File) (*RotatingWriter, error) {
	w, err := NewWriterFromFile(file)
	w.compress = true
	return w, err
}

// readCurrentSize reads the current size from the file
func (w *RotatingWriter) readCurrentSize() error {
	fi, err := w.file.Stat()
	if err != nil {
		return err
	}

	w.currentSize = fi.Size()

	return nil
}

// Daily set the rotating to be done each day.
//
// The rotating is done at (start date + 24h), not at precisely the next day.
func (w *RotatingWriter) Daily() *RotatingWriter {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.daily = true

	return w
}

// MaxSize set the size at which to rotate the file
func (w *RotatingWriter) MaxSize(s int64) *RotatingWriter {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.maxSize = s

	return w
}

// TimeFormat sets the time format to use when rolling over.
func (w *RotatingWriter) TimeFormat(s string) *RotatingWriter {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.timeFormat = s

	return w
}

// Prefix tells the writer to use the time format as prefix.
func (w *RotatingWriter) Prefix() *RotatingWriter {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.prefix = true

	return w
}

func (w *RotatingWriter) Write(b []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.daily {
		now := time.Now()
		if now.Day() != w.startDate.Day() {
			if err := w.rotate(); err != nil {
				return -1, err
			}
		}
	}

	if w.maxSize > -1 {
		if w.currentSize >= w.maxSize {
			if err := w.rotate(); err != nil {
				return -1, err
			}
		}
	}

	n, err := w.file.Write(b)
	w.currentSize += int64(n)

	return n, err
}

// rotate rotates the file. must be called while having the file lock
func (w *RotatingWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}

	{
		destName := w.makeDestName()
		_, err := os.Stat(destName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := os.Rename(w.filename, destName); err != nil {
			return err
		}

		if w.compress {
			if err := w.compressFile(destName); err != nil {
				return err
			}

			// no error to compress the data and to rename it
			// to its last filename, we can now safely remove
			// the original uncompressed file.
			if err := os.Remove(destName); err != nil {
				return err
			}
		}

		w.startDate = time.Now().Truncate(time.Hour * 24)
	}

	{
		file, err := os.OpenFile(w.filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

		w.file = file
		w.currentSize = 0
	}

	return nil
}

// compressFile compresses the file at destName into a file at destName.gz
func (w *RotatingWriter) compressFile(destName string) error {
	var rotated, tmpFile *os.File
	var err error

	// open the rotated file.
	if rotated, err = os.Open(destName); err != nil {
		return err
	}

	defer rotated.Close()

	// compress
	if tmpFile, err = w.gzip(rotated); err != nil {
		return err
	}

	defer tmpFile.Close()

	// force close just before renaming
	rotated.Close()

	// rename the gzipped file
	if err := os.Rename(tmpFile.Name(), destName+".gz"); err != nil {
		return err
	}

	return nil
}

func (w *RotatingWriter) gzip(src *os.File) (*os.File, error) {
	var tmpFile *os.File
	var err error

	// create a tmp file which will be the rotated one but compressed.
	if tmpFile, err = ioutil.TempFile(os.TempDir(), "tmp"); err != nil {
		return nil, err
	}

	// compression
	z := gzip.NewWriter(tmpFile)
	defer z.Close()
	_, err = io.Copy(z, src)
	if err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func (w *RotatingWriter) makeDestName() string {
	tf := TimeFormat
	if w.timeFormat != "" {
		tf = w.timeFormat
	}

	if w.prefix {
		ext := filepath.Ext(w.filename)
		name := w.filename[:len(w.filename)-len(ext)]

		return name + "." + w.startDate.Format(tf) + ext
	}

	return w.filename + "." + w.startDate.Format(tf)
}
