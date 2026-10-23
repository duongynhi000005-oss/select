// Package edition says which product this binary is.
package edition

// Local is set by main in a build that ships without a Select server: no
// sign-in, no sync, no proxified connections, no team administration.
var Local bool
