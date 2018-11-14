package don

import (
	"path/filepath"
	"strings"
)

func filename_add_suffix(fn string, sf string) string {
	e := filepath.Ext(fn)
	return strings.TrimRight(fn, e)+sf+e
}
