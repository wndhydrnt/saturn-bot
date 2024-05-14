package str

import "strings"

func EncloseRegex(in string) string {
	if !strings.HasPrefix(in, "^") {
		in = "^" + in
	}

	if !strings.HasSuffix(in, "$") {
		in = in + "$"
	}

	return in
}
