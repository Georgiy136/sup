package closewhmodels

type ActionCode string

const (
	ActionInProgress          ActionCode = "in_progress"
	ActionCompleted           ActionCode = "completed"
	ActionCompletedWithErrors ActionCode = "completed_with_errors"
	ActionFailed              ActionCode = "failed"
)

const (
	FailReasonLimitExceeded              = "limit_exceeded"
	FailReasonTaskError                  = "task_error"
	FailReasonAPIError                   = "api_error"
	WithErrReasonStoragePlacesNotDeleted = "storage_places_not_deleted"
)

type TicketActionResult struct {
	Action ActionCode
	Reason string
}
