package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
)

func Encrypt(ctx context.Context, text string) (string, codes.Code, error) {
	// key := []byte(keyText)
	log := logger.GetGrpcLogger(ctx)
	key := []byte(viper.GetString("encryption_key"))
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error(err)
		return "", codes.Internal, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Error(err)
		return "", codes.Internal, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext), codes.OK, nil
}

// decrypt from base64 to decrypted string
func Decrypt(ctx context.Context, cryptoText string) (string, codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	key := []byte(viper.GetString("encryption_key"))
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error(err)
		return "", codes.Internal, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	// if len(ciphertext) < aes.BlockSize {
		// err := errors.New("ciphertext too short")
		// log.Error(err)
		// return "", codes.Internal, err
	// }

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), codes.OK, nil
}
