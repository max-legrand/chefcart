/*
Package middleware ...
	Used to Encrypt / Decrypt session tokens from cookies with AES
	and verify JWT is still valid
*/
// written by: Maxwell Legrand
// tested by: Brandon Luong
// debugged by: Mark Stanik
package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	Tokens "main/webapp/protobuf"
	"main/webapp/service"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/protobuf/proto"
)

// Decrypt - decrypt token string into jwt token
func Decrypt(encryptedString string) (decryptedString string) {
	godotenv.Load(".env")
	SECRET := os.Getenv("SECRET")
	key, _ := hex.DecodeString(SECRET)
	enc, _ := hex.DecodeString(encryptedString)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Print(err.Error())
		return ""
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Print(err.Error())
		return ""
	}

	nonceSize := aesGCM.NonceSize()

	if len(enc) < nonceSize {
		return ""
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Print(err.Error())
		return ""
	}

	return fmt.Sprintf("%s", plaintext)
}

// Encrypt - encrypt jwt string into new token
func Encrypt(stringToEncrypt string) (encryptedString string) {
	godotenv.Load(".env")
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

// GenBytes - turn bytes string into array of bytes
// data = bytes string to convert to array
// separator = separator to split string by
func GenBytes(data string, separator string) []byte {
	bytes := []byte{}
	stringvals := strings.Split(data, separator)
	for index := 0; index < len(stringvals); index++ {
		value, err := strconv.Atoi(stringvals[index])
		if err != nil {
			return nil
		}
		bytes = append(bytes, byte(value))
	}
	return bytes
}

// ValidToken - determine if jwt token is valid
func ValidToken(c *gin.Context) (*jwt.Token, bool) {
	// Get token from cookie and check if valid
	prototoken := &Tokens.Token{}
	cookiedata, err := c.Cookie("token")
	if err != nil {
		return nil, false
	}
	databytes := GenBytes(cookiedata, " ")
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

// ValidTokenGRPC - determine if jwt token is valid.
// Identical in function to ValidToken
// Takes protobuf Token object as input instead of gin context pointer
func ValidTokenGRPC(tokenInput *Tokens.Token) (*jwt.Token, bool) {
	encTokenString := tokenInput.Token
	fmt.Println(encTokenString)
	databytes := GenBytes(encTokenString, "+")
	if databytes == nil || len(databytes) == 0 {
		return nil, false
	}
	newToken := &Tokens.Token{}
	proto.Unmarshal(databytes, newToken)
	tokenString := Decrypt(newToken.Token)
	if tokenString == "" {
		return nil, false
	}
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
