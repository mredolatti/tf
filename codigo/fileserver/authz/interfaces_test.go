package authz

import (
	"testing"
)

func TestPermission(t *testing.T) {
	var p IntPermission
	if res, err := p.Can(Read); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(Write); err != nil || res {
		t.Error("shouldn't be able to write.")
	}

	p.Grant(Read)
	if res, err := p.Can(Read); err != nil || !res {
		t.Error("should be able to read.")
	}

	if res, err := p.Can(Write); err != nil || res {
		t.Error("shouldn't be able to write.")
	}

	p.Grant(Write)
	if res, err := p.Can(Read); err != nil || !res {
		t.Error("should be able to read.")
	}

	if res, err := p.Can(Write); err != nil || !res {
		t.Error("should be able to write.")
	}

	p.Revoke(Read)
	if res, err := p.Can(Read); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(Write); err != nil || !res {
		t.Error("should be able to write.")
	}

	p.Revoke(Write)
	if res, err := p.Can(Read); err != nil || res {
		t.Error("shouldn't be able to read.")
	}

	if res, err := p.Can(Write); err != nil || res {
		t.Error("shouldn't be able to write.")
	}
}
