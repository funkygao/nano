// Package reqrep implements the REQ/REP protocol.  The REQ-REP socket pair
// is in lockstep. Doing any other sequence (e.g., sending two messages in
// a row) will result in error.
// REQ is the request side of the request/response pattern.
// REP is the response side of the request/response pattern.
package reqrep
