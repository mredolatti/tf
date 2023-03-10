package fsbasic

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	//"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/stretchr/testify/assert"
)

func TestFsBasicFilesMeta(t *testing.T) {

	_, err := NewFilesMetadata("/test/frula/not/exists")
	assert.NotNil(t, err)

	dir, err := ioutil.TempDir(os.TempDir(), "mifs_test")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	f, err := NewFilesMetadata(dir)
	assert.Nil(t, err)

	fm, err := f.Create("f1", "someNotes", "somePatient", "someType", time.Now().Unix())
	assert.Nil(t, err)
	assert.Equal(t, "f1", fm.ID())
	assert.Equal(t, "f1", fm.ContentID())
	assert.Equal(t, "f1", fm.Name())
	assert.Equal(t, "N/A", fm.Notes())
	assert.Equal(t, "N/A", fm.PatientID())
	assert.Equal(t, "N/A", fm.Type())
	assert.Equal(t, int64(0), fm.SizeBytes())

	fmCopy, err := f.Get("f1")
	assert.Nil(t, err)
	assert.Equal(t, fm, fmCopy)

	assert.Nil(t, ioutil.WriteFile(path.Join(dir, "f1"), []byte("hola"), 0644))
	f1WithData, err := f.Get("f1")
	assert.Nil(t, err)
	assert.Equal(t, f1WithData.SizeBytes(), int64(4))


	fm2, err := f.Create("f2", "someNotes", "somePatient", "someType", time.Now().Unix())
	assert.Nil(t, err)

	fm3, err := f.Create("f3", "someNotes", "somePatient", "someType", time.Now().Unix())
	assert.Nil(t, err)

	fms, err := f.GetMany(&apiv1.Filter{IDs: []string{"f1", "f2", "f3"}})
	assert.Nil(t, err)

	assert.Equal(t, f1WithData, fms["f1"])
	assert.Equal(t, fm2, fms["f2"])
	assert.Equal(t, fm3, fms["f3"])
}
