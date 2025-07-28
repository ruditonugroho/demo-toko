package app

// Interface PasswordHasher agar logic aplikasi bisa tetap clean dan testable
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hashed string) bool
}
