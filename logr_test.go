package logr_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vrischmann/logr"
)

func makeBuf(b byte) []byte {
	buf := make([]byte, 1024)
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

func checkEqual(t testing.TB, buf []byte, b byte) error {
	for _, v := range buf {
		if v != b {
			return fmt.Errorf("%v != %v", v, b)
		}
	}

	return nil
}

func TestRotateMaxSize(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "logr")
	require.Nil(t, err)

	rw, err := logr.NewWriterFromFile(f)
	require.Nil(t, err)

	now := time.Now()
	{
		n, err := rw.Write(makeBuf(0xFF))
		require.Nil(t, err)
		require.Equal(t, 1024, n)

		rw.MaxSize(512)

		n, err = rw.Write(makeBuf(0xFE))
		require.Nil(t, err)
		require.Equal(t, 1024, n)
	}

	newData := readFile(t, f.Name())
	require.Nil(t, checkEqual(t, newData, 0xFE))

	rotatedData := readFile(t, f.Name()+"."+now.Format(logr.TimeFormat))
	require.Nil(t, checkEqual(t, rotatedData, 0xFF))
}

func TestRotateWithCompression(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "logr")
	require.Nil(t, err)

	rw, err := logr.NewWriterFromFileWithCompression(f)
	require.Nil(t, err)

	// add some clear text at the beginning of the file
	text := []byte("This is some clear test at the beginning of the file.")

	now := time.Now()
	{
		n, err := rw.Write(text)
		require.Nil(t, err)
		require.Equal(t, len(text), n)

		n, err = rw.Write(makeBuf(0xFF))
		require.Nil(t, err)
		require.Equal(t, 1024, n)

		rw.MaxSize(512)

		n, err = rw.Write(makeBuf(0xFE))
		require.Nil(t, err)
		require.Equal(t, 1024, n)
	}

	newData := readFile(t, f.Name())
	require.Nil(t, checkEqual(t, newData, 0xFE))

	{
		f, err := os.Open(f.Name() + "." + now.Format(logr.TimeFormat) + ".gz")
		require.Nil(t, err)
		require.NotNil(t, f)

		rotatedDataGz, err := ioutil.ReadAll(f)
		require.Nil(t, err)
		require.NotEqual(t, 0, len(rotatedDataGz))

		// should not be equal cause it has been gzipped
		require.NotEqual(t, text, rotatedDataGz[:len(text)])

		// gunzip
		r, err := gzip.NewReader(bytes.NewReader(rotatedDataGz))
		require.Nil(t, err)

		gunzip, err := ioutil.ReadAll(r)
		require.Nil(t, err)

		// should be equal cause now gunzipped
		require.Equal(t, text, gunzip[:len(text)])
	}
}

func TestRotateMaxSizeCustomTimeFormat(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "logr")
	require.Nil(t, err)

	rw, err := logr.NewWriterFromFile(f)
	require.Nil(t, err)
	rw.TimeFormat("2006__01__02")

	now := time.Now()
	{
		n, err := rw.Write(makeBuf(0xFF))
		require.Nil(t, err)
		require.Equal(t, 1024, n)

		rw.MaxSize(512)

		n, err = rw.Write(makeBuf(0xFE))
		require.Nil(t, err)
		require.Equal(t, 1024, n)
	}

	newData := readFile(t, f.Name())
	require.Nil(t, checkEqual(t, newData, 0xFE))

	rotatedData := readFile(t, f.Name()+"."+now.Format("2006__01__02"))
	require.Nil(t, checkEqual(t, rotatedData, 0xFF))
}

func TestRotateMaxSizePrefix(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "logr")
	require.Nil(t, err)

	rw, err := logr.NewWriterFromFile(f)
	require.Nil(t, err)
	rw.Prefix()

	now := time.Now()
	{
		n, err := rw.Write(makeBuf(0xFF))
		require.Nil(t, err)
		require.Equal(t, 1024, n)

		rw.MaxSize(512)

		n, err = rw.Write(makeBuf(0xFE))
		require.Nil(t, err)
		require.Equal(t, 1024, n)
	}

	newData := readFile(t, f.Name())
	require.Nil(t, checkEqual(t, newData, 0xFE))

	ext := filepath.Ext(f.Name())
	name := f.Name()[:len(f.Name())-len(ext)]
	rotatedData := readFile(t, name+"."+now.Format(logr.TimeFormat)+ext)
	require.Nil(t, checkEqual(t, rotatedData, 0xFF))
}
