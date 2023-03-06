package fsbasic

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/stretchr/testify/assert"
)

func TestFsBasicFiles(t *testing.T) {

	_, err := NewFiles("/test/frula/not/exists")
	assert.NotNil(t, err)

	dir, err := ioutil.TempDir(os.TempDir(), "mifs_test")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	f, err := NewFiles(dir)
	assert.Nil(t, err)

	fileId := "someId"
	_, err = f.Read(fileId)
	assert.ErrorIs(t, err, apiv1.ErrFileDoesNotExist)

	err = f.Write(fileId, []byte("some_data"), false)
	assert.Nil(t, err)

	data, err := f.Read(fileId)
	assert.Nil(t, err)
	assert.Equal(t, []byte("some_data"), data)

	err = f.Write(fileId, []byte("some_data"), false)
	assert.ErrorIs(t, err,apiv1.ErrFileExists)

	err = f.Write(fileId, []byte("some_data2"), true)
	assert.Nil(t, err)
	data, err = f.Read(fileId)
	assert.Nil(t, err)
	assert.Equal(t, []byte("some_data2"), data)
}
