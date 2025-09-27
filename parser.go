package golog

import (
	"strings"
)

// mergeMaps copies entries from the provided maps into dst. Keys are normalized
// inline (trim spaces, remove trailing colon, strip surrounding quotes).
func mergeMaps(dst map[string]any, maps ...map[string]any) {
	for _, m := range maps {
		if m == nil {
			continue
		}
		for k, v := range m {
			// skip normalization when the key
			// appears to be already normalized (no surrounding quotes,
			// no trailing colon, and no leading/trailing space).
			if !strings.ContainsAny(k, " \t\n\r:'\"") {
				dst[k] = v
				continue
			}

			nk := strings.TrimSpace(k)
			nk = strings.TrimSuffix(nk, ":")
			nk = strings.Trim(nk, "\"'")
			nk = strings.TrimSpace(nk)
			dst[nk] = v
		}
	}
}
