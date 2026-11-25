package model

import "time"

// Nama event yang ditulis producer ke tabel outbox.
const EventAppointmentCreated = "appointment.created"

// Nama jenis task dan queue Asynq yang dipakai worker notifikasi.
const (
	TaskSendNotification = "notification:send"
	QueueNotifications   = "notifications"
)

// OutboxEvent adalah satu baris outbox yang belum dipublish.
type OutboxEvent struct {
	ID        int64
	EventType string
	Payload   []byte
	CreatedAt time.Time
}

// AppointmentCreatedEvent adalah kontrak payload event appointment.created.
// Field ini disalin dari data booking saat event ditulis, bukan diambil lagi
// dari database agar notifikasi tetap sesuai dengan keadaan saat booking.
type AppointmentCreatedEvent struct {
	UserID          int64  `json:"user_id"`
	AppointmentID   int64  `json:"appointment_id"`
	AppointmentDate string `json:"appointment_date"`
	QueueNumber     int    `json:"queue_number"`
	PatientName     string `json:"patient_name"`
	DoctorName      string `json:"doctor_name"`
}

// NotificationJob adalah isi task yang dikirim relay ke Asynq.
// OutboxID adalah idempotency key untuk worker.
type NotificationJob struct {
	OutboxID  int64  `json:"outbox_id"`
	EventType string `json:"event_type"`
	Payload   []byte `json:"payload"`
}

// Message adalah pesan yang akan dikirim mailer.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Notification adalah bukti terkirimnya satu event ke seorang user.
type Notification struct {
	ID       int64     `json:"id"`
	OutboxID int64     `json:"outboxId"`
	UserID   int64     `json:"userId"`
	Channel  string    `json:"channel"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
	SentAt   time.Time `json:"sentAt"`
}
