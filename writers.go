package bcl

import "io"

type writers struct {
	outw, logw io.Writer
}
