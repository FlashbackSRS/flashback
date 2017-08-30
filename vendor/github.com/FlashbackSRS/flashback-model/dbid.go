package fb

import (
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var validDBIDTypes = map[string]struct{}{
	"user":   {},
	"bundle": {},
}

// Same as standard Base32 encoding, only lowercase to work with CouchDB database
// naming restrictions.
var b32encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")

// B32enc encodes the input using Flashback's Base32 variant.
func B32enc(data []byte) string {
	return strings.TrimRight(b32encoding.EncodeToString(data), "=")
}

// B32dec decodes Flashback's Base32 variant.
func B32dec(s string) ([]byte, error) {
	// fmt.Printf("Before: '%s'\n", s)
	if padLen := len(s) % 8; padLen > 0 {
		s = s + strings.Repeat("=", 8-padLen)
	}
	return b32encoding.DecodeString(s)
}

func validateDBID(id string) error {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return errors.New("invalid DBID format")
	}
	if _, ok := validDBIDTypes[parts[0]]; !ok {
		return errors.Errorf("unsupported DBID type '%s'", parts[0])
	}
	if _, err := B32dec(parts[1]); err != nil {
		return errors.New("invalid DBID encoding")
	}
	return nil
}

// EncodeDBID generates a DBID by encoding the docType and Base32-encoding
// the ID. No validation is done of the docType.
func EncodeDBID(docType string, id []byte) string {
	return fmt.Sprintf("%s-%s", docType, B32enc(id))
}
