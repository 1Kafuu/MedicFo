package model

// AppointmentReport adalah satu baris laporan antrean harian.
// CheckedInAt sudah berupa teks dari database agar formatnya sama antara
// respons JSON dan berkas CSV.
type AppointmentReport struct {
	QueueNumber int    `json:"queueNumber"`
	PatientName string `json:"patientName"`
	DoctorName  string `json:"doctorName"`
	Status      string `json:"status"`
	CheckedInAt string `json:"checkedInAt"`
}
