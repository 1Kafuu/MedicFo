package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost: berapa kali proses internal hashing diulang. Nilai 12 dipakai
const bcryptCost = 12

var dummyHash = func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("medicfo-dummy-password"), bcryptCost)
	if err != nil {
		panic(err) // tidak mungkin: bcrypt dari string tetap
	}
	return h
}()

// HashPassword mengubah password menjadi hash yang tidak dapat dikembalikan.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword membandingkan password dengan hash-nya.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// VerifyDummyPassword membuang waktu seperti VerifyPassword
func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}

// RandomToken menghasilkan string acak yang aman secara kriptografis.
func RandomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SHA256Hex dipakai untuk menyimpan refresh token dalam bentuk hash.
func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
