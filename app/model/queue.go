package model

// QueueStats adalah ringkasan keadaan outbox dan queue untuk admin.
// QueueError diisi bila Redis tidak dapat dibaca; statistik PostgreSQL
// tetap dikirim karena itulah data yang paling dibutuhkan saat Redis mati.
type QueueStats struct {
	OutboxPending     int        `json:"outbox_pending"`
	OutboxPublished   int        `json:"outbox_published"`
	OldestPendingAge  string     `json:"oldest_pending_age"`
	NotificationsSent int        `json:"notifications_sent"`
	Queue             *QueueInfo `json:"queue,omitempty"`
	QueueError        string     `json:"queue_error,omitempty"`
}

// QueueInfo adalah salinan sebagian asynq.QueueInfo. Sengaja ditulis ulang
// agar package model tidak bergantung pada library eksternal dan bentuk
// respons API tidak ikut berubah bila Asynq mengubah struct-nya.
type QueueInfo struct {
	Name      string `json:"name"`
	Pending   int    `json:"pending"`
	Active    int    `json:"active"`
	Scheduled int    `json:"scheduled"`
	Retry     int    `json:"retry"`
	Archived  int    `json:"archived"`
	Completed int    `json:"completed"`
	Processed int    `json:"processed_today"`
	Failed    int    `json:"failed_today"`
}
