package fsbasic

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/stretchr/testify/assert"
)

func TestFsBasicAuthz(t *testing.T) {

	dir, err := ioutil.TempDir(os.TempDir(), "authz_test")
	assert.Nil(t, err)

	authz, err := NewAuthz(dir)
	assert.Nil(t, err)
	defer authz.db.Close()

	can, err := authz.Can("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)
	assert.False(t, can)

	err = authz.Grant("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)

	can, err = authz.Can("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)
	assert.True(t, can)

	err = authz.Revoke("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)

	can, err = authz.Can("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)
	assert.False(t, can)

	err = authz.Grant("martin", apiv1.OperationWrite, "file1.txt")
	assert.Nil(t, err)
	err = authz.Grant("martin", apiv1.OperationWrite, "file2.txt")
	assert.Nil(t, err)
	err = authz.Grant("martin", apiv1.OperationWrite, "file3.txt")
	assert.Nil(t, err)
}
