/*
Copyright © 2025 srz_zumix
*/
package profile

import (
	"fmt"
	"slices"

	"github.com/google/go-github/v90/github"
)

// Field identifies a single profile attribute that can be copied.
type Field string

const (
	FieldName            Field = "name"
	FieldBio             Field = "bio"
	FieldCompany         Field = "company"
	FieldLocation        Field = "location"
	FieldBlog            Field = "blog"
	FieldTwitterUsername Field = "twitter_username"
	FieldHireable        Field = "hireable"
	FieldSocialAccounts  Field = "social_accounts"
)

// AllFields returns every field supported by "profile copy", in a stable order.
func AllFields() []string {
	return []string{
		string(FieldName),
		string(FieldBio),
		string(FieldCompany),
		string(FieldLocation),
		string(FieldBlog),
		string(FieldTwitterUsername),
		string(FieldHireable),
		string(FieldSocialAccounts),
	}
}

// ParseFields validates the requested field names against AllFields and
// returns them unchanged. An empty input is not expanded here; callers should
// pass AllFields() to copy everything.
func ParseFields(fields []string) ([]string, error) {
	valid := AllFields()
	for _, f := range fields {
		if !slices.Contains(valid, f) {
			return nil, fmt.Errorf("unknown field %q (valid fields: %v)", f, valid)
		}
	}
	return fields, nil
}

// BuildUpdateRequest returns a *github.User containing only the fields
// present in both fields and src (src fields left empty/nil are skipped so
// the destination's existing value is preserved), along with the list of
// field names that will actually be changed.
func BuildUpdateRequest(src *github.User, fields []string) (*github.User, []string) {
	req := &github.User{}
	var changed []string
	for _, f := range fields {
		switch Field(f) {
		case FieldName:
			if src.Name != nil {
				req.Name = src.Name
				changed = append(changed, f)
			}
		case FieldBio:
			if src.Bio != nil {
				req.Bio = src.Bio
				changed = append(changed, f)
			}
		case FieldCompany:
			if src.Company != nil {
				req.Company = src.Company
				changed = append(changed, f)
			}
		case FieldLocation:
			if src.Location != nil {
				req.Location = src.Location
				changed = append(changed, f)
			}
		case FieldBlog:
			if src.Blog != nil {
				req.Blog = src.Blog
				changed = append(changed, f)
			}
		case FieldTwitterUsername:
			if src.TwitterUsername != nil {
				req.TwitterUsername = src.TwitterUsername
				changed = append(changed, f)
			}
		case FieldHireable:
			if src.Hireable != nil {
				req.Hireable = src.Hireable
				changed = append(changed, f)
			}
		case FieldSocialAccounts:
			// handled separately via MissingSocialAccounts
		}
	}
	return req, changed
}

// MissingSocialAccounts returns the URLs of social accounts present in src
// but not in dst.
func MissingSocialAccounts(src, dst []*github.SocialAccount) []string {
	dstURLs := make(map[string]bool, len(dst))
	for _, a := range dst {
		dstURLs[a.GetURL()] = true
	}
	var missing []string
	for _, a := range src {
		if url := a.GetURL(); url != "" && !dstURLs[url] {
			missing = append(missing, url)
		}
	}
	return missing
}
