package demo

type Status int

const (
	DefaultRetryLimit = 3
	DefaultPageSize   = 20
)

const (
	StatusUnknown Status = iota
	StatusReady
	StatusFailed
)
