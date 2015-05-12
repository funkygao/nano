// Package pipeline implements a one-way PUSH/PULL protocol.
// PULL protocol is the reader side of the pipeline pattern.
// PUSH protocol is the writer side of the pipeline pattern.
// If a PUSHer has multiple PULLers, a single message will be load
// balanced to one of the PULLers.
package pipeline
