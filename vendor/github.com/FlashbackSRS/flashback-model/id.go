package fb

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var validDocIDTypes = map[string]struct{}{
	"theme": {},
	"note":  {},
	"deck":  {},
	"card":  {},
}

func validateDocID(id string) error {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return errors.New("invalid DocID format")
	}
	if _, ok := validDocIDTypes[parts[0]]; !ok {
		return errors.Errorf("unsupported DocID type '%s'", parts[0])
	}
	if _, err := b64encoder.DecodeString(parts[1]); err != nil {
		return errors.New("invalid DocID encoding")
	}
	return nil
}

// EncodeDocID generates a DocID by encoding the docType and Base64-encoding
// the ID. No validation is done of the docType.
func EncodeDocID(docType string, id []byte) string {
	return fmt.Sprintf("%s-%s", docType, b64encoder.EncodeToString(id))
}
