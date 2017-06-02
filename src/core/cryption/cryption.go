package cryption

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	COOKIE_JWT_MAP            = "jwt-cookie"
	JWT_EXPIRY_TIME           = 10 * time.Minute
	JWT_EXPIRY_TIME_TEST_FAIL = 1 * time.Second
	JWT_EXPIRY_TIME_NEW_PASS  = 10 * time.Minute
)

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func NewJWTString(email string, hmacSecret [32]byte, waitTime time.Duration) (string, error) {
	loc, _ := time.LoadLocation("UTC")

	claims := jws.Claims{
		"email": email,
	}
	claims.SetNotBefore(time.Now().In(loc))
	claims.SetExpiration(time.Now().In(loc).Add(waitTime))

	token := jws.NewJWT(claims, crypto.SigningMethodHS256)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.Serialize(hmacSecret[:])
	return string(tokenString), err
}

func VerifyJWTGetEmail(tokenString string, hmacSecret [32]byte) (string, error) {
	token, err := VerifyJWT(tokenString, hmacSecret)
	if err != nil {
		return "", err
	}
	email, ok := token.Claims().Get("email").(string)
	if !ok {
		return "", fmt.Errorf("Error Retrieving email from JWT: %s\n", err.Error())
	}
	return email, nil
}

func VerifyJWT(tokenString string, hmacSecret [32]byte) (jwt.JWT, error) {
	token, err := jws.ParseJWT([]byte(tokenString))
	if err != nil {
		return nil, fmt.Errorf("ERROR Failed Verify JWT %s", err.Error())
	}
	if err := token.Validate(hmacSecret[:], crypto.SigningMethodHS256); err != nil {
		return nil, fmt.Errorf("ERROR Verifying JWT: %s\n", err.Error())
	}
	return token, nil
}

func ParseJWT(tokenString string) (jwt.JWT, error) {
	return jws.ParseJWT([]byte(tokenString))
}

func GetJWTSignature(tokenString string) (string, error) {
	arr := strings.Split(tokenString, ".")
	if len(arr) != 3 {
		return "", fmt.Errorf("Error with jwt, not proper size: %s\n", len(arr))
	}
	return arr[2], nil
}
