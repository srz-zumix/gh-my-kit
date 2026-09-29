/*
Copyright © 2025 srz_zumix
*/
package profile

import (
	"testing"

	"github.com/google/go-github/v90/github"
)

func TestParseFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{"all valid", AllFields(), false},
		{"single valid", []string{"name"}, false},
		{"empty", []string{}, false},
		{"unknown", []string{"avatar"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFields(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFields(%v) error = %v, wantErr %v", tt.fields, err, tt.wantErr)
			}
		})
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	src := &github.User{
		Name:     github.Ptr("Alice"),
		Bio:      github.Ptr("hello"),
		Location: github.Ptr(""),
	}

	req, changed := BuildUpdateRequest(src, []string{"name", "bio", "location", "company"})

	if req.GetName() != "Alice" {
		t.Errorf("Name = %q, want Alice", req.GetName())
	}
	if req.GetBio() != "hello" {
		t.Errorf("Bio = %q, want hello", req.GetBio())
	}
	if req.Company != nil {
		t.Errorf("Company should stay nil, got %q", req.GetCompany())
	}
	if req.Location != nil {
		t.Errorf("empty Location should be skipped, got %q", req.GetLocation())
	}

	wantChanged := []string{"name", "bio"}
	if len(changed) != len(wantChanged) {
		t.Fatalf("changed = %v, want %v", changed, wantChanged)
	}
	for i, f := range wantChanged {
		if changed[i] != f {
			t.Errorf("changed[%d] = %q, want %q", i, changed[i], f)
		}
	}
}

func TestBuildUpdateRequest_SkipsNilFields(t *testing.T) {
	src := &github.User{}
	req, changed := BuildUpdateRequest(src, AllFields())

	if len(changed) != 0 {
		t.Errorf("changed = %v, want empty", changed)
	}
	if req.Name != nil || req.Bio != nil || req.Hireable != nil {
		t.Errorf("expected all fields to remain nil, got %+v", req)
	}
}

func TestBuildUpdateRequest_HireableFalse(t *testing.T) {
	src := &github.User{Hireable: github.Ptr(false)}
	req, changed := BuildUpdateRequest(src, []string{"hireable"})

	if req.Hireable == nil || req.GetHireable() != false {
		t.Errorf("Hireable = %v, want false (non-nil)", req.Hireable)
	}
	if len(changed) != 1 || changed[0] != "hireable" {
		t.Errorf("changed = %v, want [hireable]", changed)
	}
}

func TestMissingSocialAccounts(t *testing.T) {
	src := []*github.SocialAccount{
		{URL: github.Ptr("https://twitter.com/alice")},
		{URL: github.Ptr("https://github.com/alice")},
	}
	dst := []*github.SocialAccount{
		{URL: github.Ptr("https://github.com/alice")},
	}

	missing := MissingSocialAccounts(src, dst)
	if len(missing) != 1 || missing[0] != "https://twitter.com/alice" {
		t.Errorf("missing = %v, want [https://twitter.com/alice]", missing)
	}
}

func TestMissingSocialAccounts_NoneMissing(t *testing.T) {
	accounts := []*github.SocialAccount{
		{URL: github.Ptr("https://github.com/alice")},
	}
	missing := MissingSocialAccounts(accounts, accounts)
	if len(missing) != 0 {
		t.Errorf("missing = %v, want empty", missing)
	}
}
