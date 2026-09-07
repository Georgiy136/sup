package bandaction

type BandActionService struct {
	ticketRepo      ticketActionRepo
	bandBot         bandBotClient
	ticketPostCache ticketPostCache
	actionSigner    actionSigner
}

func NewBandActionService(
	ticketRepo ticketActionRepo,
	bandBot bandBotClient,
	ticketPostCache ticketPostCache,
	actionSigner actionSigner,
) *BandActionService {
	return &BandActionService{
		ticketRepo:      ticketRepo,
		bandBot:         bandBot,
		ticketPostCache: ticketPostCache,
		actionSigner:    actionSigner,
	}
}
