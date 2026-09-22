package activity

import (
	"fmt"
	"strings"
)

// ResolveKinds computes the effective set of kinds to collect from --include/--exclude
// flag values. include and exclude are mutually exclusive. Unknown kind names are rejected.
func ResolveKinds(include, exclude []string) ([]Kind, error) {
	if len(include) > 0 && len(exclude) > 0 {
		return nil, fmt.Errorf("--include cannot be used together with --exclude")
	}

	valid := make(map[Kind]bool, len(AllKinds))
	for _, k := range AllKinds {
		valid[k] = true
	}
	toKinds := func(names []string) ([]Kind, error) {
		kinds := make([]Kind, 0, len(names))
		for _, name := range names {
			k := Kind(strings.TrimSpace(name))
			if !valid[k] {
				return nil, fmt.Errorf("unknown activity kind %q", name)
			}
			kinds = append(kinds, k)
		}
		return kinds, nil
	}

	if len(include) > 0 {
		return toKinds(include)
	}
	if len(exclude) > 0 {
		excluded, err := toKinds(exclude)
		if err != nil {
			return nil, err
		}
		excludedSet := make(map[Kind]bool, len(excluded))
		for _, k := range excluded {
			excludedSet[k] = true
		}
		kinds := make([]Kind, 0, len(AllKinds))
		for _, k := range AllKinds {
			if !excludedSet[k] {
				kinds = append(kinds, k)
			}
		}
		return kinds, nil
	}

	return append([]Kind(nil), AllKinds...), nil
}
