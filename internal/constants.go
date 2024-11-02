package internal

import "fmt"

const NAME = "readflow"

type ReadStatus int

const (
	STATUS_UNREAD ReadStatus = iota
	STATUS_FINISHED
	STATUS_IN_PROGRESS
)

func (e ReadStatus) String() string {
	switch e {
	case STATUS_UNREAD:
		return "UNREAD"
	case STATUS_FINISHED:
		return "FINISHED"
	case STATUS_IN_PROGRESS:
		return "IN_PROGRESS"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}
