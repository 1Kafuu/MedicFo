package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadENV memuat variabel dari berkas .env ke environment proses.
// Bila berkasnya tidak ada, aplikasi tetap jalan memakai environment sistem.
func LoadENV() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: berkas .env tidak ditemukan, memakai environment sistem")
	}
}

// GetENV mengambil nilai environment. Bila kosong, kembalikan nilai bawaan.
func GetENV(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// GetENVInt sama seperti GetENV, tetapi hasilnya berupa int.
func GetENVInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Peringatan: %s bukan angka (%q), memakai bawaan %d", key, value, fallback)
		return fallback
	}
	return parsed
}