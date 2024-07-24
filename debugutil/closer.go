package debugutil

import (
	"io"
)

func SafeClose(closer io.Closer) {
	if closer != nil {
		_ = Verify(closer.Close())
	}
}
