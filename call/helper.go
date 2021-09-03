package call

import (
	"strings"

	guuid "github.com/google/uuid"
)

// GenerateCallSID generates the call SID with unique 32 length hexadecimal
func GenerateCallSID() string {
	id := guuid.New().String()
	callSID := strings.Replace(id, "-", "", -1)
	return callSID
}
