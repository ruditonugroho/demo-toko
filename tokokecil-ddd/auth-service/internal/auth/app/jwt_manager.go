package app

// Interface JWTManager supaya bisa diganti2 implementasi tanpa ubah logic aplikasi
type JWTManager interface {
	GenerateToken(userID, email string) (string, error)
}
