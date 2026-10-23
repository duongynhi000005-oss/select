// Package edition says which product this binary is.
package edition

// name is set at link time: -X selectDb/internal/edition.name=local.
var name string

// Local is true in a build that ships without a Select server: no sign-in, no
// sync, no proxified connections, no team administration.
var Local = name == "local"
