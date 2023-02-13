package authz

import (
	"testing"
)

func TestPermission(t *testing.T) {
	var p IntPermission
	if res, err := p.Can(OperationRead); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(OperationWrite); err != nil || res {
		t.Error("shouldn't be able to write.")
	}

	p.Grant(OperationRead)
	if res, err := p.Can(OperationRead); err != nil || !res {
		t.Error("should be able to read.")
	}

	if res, err := p.Can(OperationWrite); err != nil || res {
		t.Error("shouldn't be able to write.")
	}

	p.Grant(OperationWrite)
	if res, err := p.Can(OperationRead); err != nil || !res {
		t.Error("should be able to read.")
	}

	if res, err := p.Can(OperationWrite); err != nil || !res {
		t.Error("should be able to write.")
	}

	p.Revoke(OperationRead)
	if res, err := p.Can(OperationRead); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(OperationWrite); err != nil || !res {
		t.Error("should be able to write.")
	}

	p.Revoke(OperationWrite)
	if res, err := p.Can(OperationRead); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(OperationWrite); err != nil || res {
		t.Error("shouldn't be able to write.")
	}
}
