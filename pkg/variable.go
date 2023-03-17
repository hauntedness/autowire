package pkg

import "go/types"

var (
	errorType   = types.Universe.Lookup("error").Type()
	cleanupType = types.NewSignatureType(nil, nil, nil, nil, nil, false)
)
