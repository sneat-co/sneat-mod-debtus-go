package general

type CreatedOn struct {
	CreatedOnPlatform string `datastore:",noindex"` // e.g. "Telegram"
	CreatedOnID       string `datastore:",noindex"` // e.g. "DebtsTrackerBot"
}
