package cryption

import (
	"github.com/dgrijalva/jwt-go"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	COOKIE_JWT_MAP            = "jwt-cookie"
	JWT_EXPIRY_TIME           = 1 * time.Hour
	JWT_EXPIRY_TIME_TEST_FAIL = 5 * time.Minute
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

func NewJWT(email string, hmacSecret [32]byte, waitTime time.Duration) (tokenString string, err error) {
	loc, _ := time.LoadLocation("UTC")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": email,
		"nbf": time.Now().In(loc).Unix(),
		"exp": time.Now().In(loc).Add(waitTime).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(hmacSecret[:])
}

func VerifyJWT(tokenString string, hmacSecret [32]byte) (email string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSecret is a []byte containing your secret
		return hmacSecret[:], nil
	})

	if err != nil {
		return "", fmt.Errorf("ERROR Failed Verify JWT %s", err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["uid"], claims["nbf"], claims["exp"])
		return claims["uid"].(string), nil
	} else {
		fmt.Printf("ERROR Verifying JWT: %s\n", err.Error())
		return "", err
	}
}
