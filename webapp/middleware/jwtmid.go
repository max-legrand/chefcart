/*
Package middleware ...
	Used to Encrypt / Decrypt session tokens from cookies with AES
	and verify JWT is still valid
*/
package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"main/webapp/protobuf"
	"main/webapp/service"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/protobuf/proto"
)

// Decrypt ...
func Decrypt(encryptedString string) (decryptedString string) {
	godotenv.Load("../.env")
	key, _ := hex.DecodeString(os.Getenv("SECRET"))
	enc, _ := hex.DecodeString(encryptedString)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonceSize := aesGCM.NonceSize()

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}

// Encrypt ...
func Encrypt(stringToEncrypt string) (encryptedString string) {
	godotenv.Load("../.env")
	key, _ := hex.DecodeString(os.Getenv("SECRET"))
	plaintext := []byte(stringToEncrypt)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

// GenBytes ...
func GenBytes(data string) []byte {
	bytes := []byte{}
	stringvals := strings.Split(data, " ")
	for index := 0; index < len(stringvals); index++ {
		value, err := strconv.Atoi(stringvals[index])
		if err != nil {
			return nil
		}
		bytes = append(bytes, byte(value))
	}
	return bytes
}

// ValidToken ...
func ValidToken(c *gin.Context) (*jwt.Token, bool) {
	// Get token from cookie and check if valid
	prototoken := &protobuf.Token{}
	cookiedata, err := c.Cookie("token")
	if err != nil {
		return nil, false
	}
	databytes := GenBytes(cookiedata)
	if databytes == nil || len(databytes) == 0 {
		return nil, false
	}
	err = proto.Unmarshal(databytes, prototoken)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	encTokenString := prototoken.Token
	fmt.Println(encTokenString)
	tokenString := Decrypt(encTokenString)
	fmt.Println(tokenString)
	token, err := service.JWTAuthService().ValidateToken(tokenString)
	if token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		fmt.Println(claims)
		return token, true
	}
	fmt.Println(err)
	return nil, false

}
