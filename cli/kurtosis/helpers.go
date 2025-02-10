package kurtosis

import (
	"go.starlark.net/starlark"
)

func MergeDicts(dicts ...starlark.StringDict) starlark.StringDict {
	var merged = map[string]starlark.Value{}

	for _, dict := range dicts {
		for k, v := range dict {
			merged[k] = v
		}
	}

	return merged
}
