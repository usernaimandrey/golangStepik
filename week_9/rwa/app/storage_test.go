package app

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestUserUpdate(t *testing.T) {
	user := &User{
		Email:    "test@test",
		UserName: "test",
	}

	userData := &User{
		Email: "update@test",
	}

	expected := &User{
		Email:    "update@test",
		UserName: "test",
	}

	user.Update(userData)

	if diff := cmp.Diff(user, expected); diff != "" {
		t.Error(diff)
	}
}
