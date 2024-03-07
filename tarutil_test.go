package tarutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	for _, src := range []string{"./testdata/hello.txt", "./testdata/test"} {
		for _, dst := range []string{"./testdata/test.tar", "./testdata/test.tar.gz"} {
			err := Archive(src, dst)
			require.NoError(t, err)

			require.Equal(t, true, pathExists(dst))
			require.NoError(t, os.Remove(dst))
		}
	}
}

func TestExtract(t *testing.T) {
	src, outdir, dst := "./testdata/hello.txt", "./testdata/outdir", "./testdata/test.tar.gz"

	err := Archive(src, dst)
	require.NoError(t, err)

	err = Extract(dst, outdir)
	require.NoError(t, err)

	d1, err := ioutil.ReadFile(src)
	require.NoError(t, err)

	d2, err := ioutil.ReadFile(filepath.Join(outdir, "hello.txt"))
	require.NoError(t, err)

	require.Equal(t, d1, d2)

	require.NoError(t, os.RemoveAll(outdir))
	require.NoError(t, os.Remove(dst))
}
