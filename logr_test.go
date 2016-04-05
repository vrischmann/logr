package logr

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func makeBuf(b byte, n int) []byte {
	buf := make([]byte, n)
	for i := 0; i < len(buf); i++ {
		buf[i] = b
	}

	return buf
}

func readFile(t testing.TB, filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	require.Nil(t, err)

	return data
}

func makeTempFile(t testing.TB) *os.File {
	f, err := ioutil.TempFile("", "logr_test")
	require.Nil(t, err)

	return f
}

func checkEqual(buf []byte, b byte) error {
	for _, v := range buf {
		if v != b {
			return fmt.Errorf("%v != %v", v, b)
		}
	}

	return nil
}

func removeFile(t testing.TB, f *os.File, w *RotatingWriter) {
	require.Nil(t, w.file.Close())
	require.Nil(t, os.Remove(f.Name()))
}

func TestRotateDaily(t *testing.T) {
	f := makeTempFile(t)
	opts := &Options{RotateDaily: true}

	w, err := NewWriterFromFile(f, opts)
	require.Nil(t, err)

	defer removeFile(t, f, w)

	// first write - will always write to the original file
	n, err := w.Write(makeBuf('A', 80))
	require.Nil(t, err)
	require.Equal(t, 80, n)

	// force the last mod date to yesterday to trigger a rotation
	w.lastMod = w.lastMod.Add(-25 * time.Hour)

	rotatedName := makeDestName(f.Name(), w.lastMod, opts)
	defer os.Remove(rotatedName)

	// second write - will write to the new file because time.Now() - w.lastMod < 24h
	n, err = w.Write(makeBuf('B', 30))
	require.Nil(t, err)
	require.Equal(t, 30, n)

	// recheck the rotated file to see if it changed
	data := readFile(t, rotatedName)
	require.Equal(t, 80, len(data))
	require.Nil(t, checkEqual(data, 'A'))

	// check the new file
	data = readFile(t, f.Name())
	require.Equal(t, 30, len(data))
	require.Nil(t, checkEqual(data, 'B'))
}

func TestRotateMaximumSize(t *testing.T) {
	f := makeTempFile(t)
	opts := &Options{MaximumSize: 100}

	w, err := NewWriterFromFile(f, opts)
	require.Nil(t, err)

	defer removeFile(t, f, w)

	// first write - will always write to the original file
	n, err := w.Write(makeBuf('A', 80))
	require.Nil(t, err)
	require.Equal(t, 80, n)

	data := readFile(t, f.Name())
	require.Equal(t, 80, len(data))
	require.Nil(t, checkEqual(data, 'A'))

	// second write - will write to the original file because currentSize (80) < maximum size (100)
	n, err = w.Write(makeBuf('B', 30))
	require.Nil(t, err)
	require.Equal(t, 30, n)

	data = readFile(t, f.Name())
	require.Equal(t, 110, len(data))
	require.Nil(t, checkEqual(data[80:], 'B'))

	rotatedName := makeDestName(f.Name(), getMidnightFromDate(time.Now()), opts)
	defer os.Remove(rotatedName)

	// third write - will trigger a rotation
	n, err = w.Write(makeBuf('C', 50))
	require.Nil(t, err)
	require.Equal(t, 50, n)

	// recheck the rotated file to see if it changed
	data = readFile(t, rotatedName)
	require.Equal(t, 110, len(data))
	require.Nil(t, checkEqual(data[:80], 'A'))
	require.Nil(t, checkEqual(data[80:], 'B'))

	// check the new file
	data = readFile(t, f.Name())
	require.Equal(t, 50, len(data))
	require.Nil(t, checkEqual(data, 'C'))
}

func TestRotateWithCompression(t *testing.T) {
	f := makeTempFile(t)
	opts := &Options{
		MaximumSize: 100,
		Compress:    true,
	}

	w, err := NewWriterFromFile(f, opts)
	require.Nil(t, err)

	defer removeFile(t, f, w)

	// first write - will always write to the original file
	n, err := w.Write(makeBuf('A', 80))
	require.Nil(t, err)
	require.Equal(t, 80, n)

	data := readFile(t, f.Name())
	require.Equal(t, 80, len(data))
	require.Nil(t, checkEqual(data, 'A'))

	// second write - will write to the original file because currentSize (80) < maximum size (100)
	n, err = w.Write(makeBuf('B', 30))
	require.Nil(t, err)
	require.Equal(t, 30, n)

	data = readFile(t, f.Name())
	require.Equal(t, 110, len(data))
	require.Nil(t, checkEqual(data[80:], 'B'))

	rotatedName := makeDestName(f.Name(), getMidnightFromDate(time.Now()), opts)

	// third write - will trigger a rotation
	n, err = w.Write(makeBuf('C', 50))
	require.Nil(t, err)
	require.Equal(t, 50, n)

	// check the new file
	data = readFile(t, f.Name())
	require.Equal(t, 50, len(data))
	require.Nil(t, checkEqual(data, 'C'))

	// check the gzipped file
	time.Sleep(1 * time.Second)

	gzName := rotatedName + ".gz"
	data = readFile(t, gzName)
	defer os.Remove(gzName)

	gr, err := gzip.NewReader(bytes.NewReader(data))
	require.Nil(t, err)

	data, err = ioutil.ReadAll(gr)
	require.Nil(t, err)

	require.Equal(t, 110, len(data))
	require.Nil(t, checkEqual(data[:80], 'A'))
	require.Nil(t, checkEqual(data[80:], 'B'))
}
