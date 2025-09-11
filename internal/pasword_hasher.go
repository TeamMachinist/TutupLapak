package internal

// TODO: Experimental change to Argon2
import "golang.org/x/crypto/bcrypt"

type PasswordHasher interface {
	EncryptPassword(password string) (string, err)
	ComparePasswordHash(passwordHash string, plainPassword string) err
}

type passwordHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return &passwordHasher{}
}

func (ph *passwordHasher) EncryptPassword(password string) (string, err) {
	passwordHash, error := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}

func (ph *passwordHasher) ComparePasswordHash(passwordHash string, plainPassword string) err {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plainPassword))
}
