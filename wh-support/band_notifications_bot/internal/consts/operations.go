package consts

const (
	ApproveOperation   = "approve"   // Заявка требует подтверждения
	ApprovedOperation  = "approved"  // Заявка подтверждена пользователем
	PerformOperation   = "perform"   // Заявка требует исполнения
	PerformedOperation = "performed" // Заявка исполнена пользователем
	BookedOperation    = "booked"    // Заявка взята в работу пользователем
	UnbookedOperation  = "unbooked"  // Заявка снята с работы пользователем
	RejectOperation    = "rejected"  // Заявка отклонена пользователем
	CompletedOperation = "completed" // Заявка выполнена пользователем
	ReturnedOperation  = "returned"  // Заявка возвращена на предыдущий статус
)
