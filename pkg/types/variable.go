package types

import "go/types"

var (
	errorType   = types.Universe.Lookup("error").Type()
	cleanupType = types.NewSignature(nil, nil, nil, false)
)
