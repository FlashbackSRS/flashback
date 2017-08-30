package fb

import (
	"encoding/base64"
)

var b64encoder = base64.URLEncoding.WithPadding(base64.NoPadding)
