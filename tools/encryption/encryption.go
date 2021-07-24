package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
)

var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(ctx context.Context, s string) ([]byte, codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Error(err)
		return nil, codes.Internal, err
	}
	return data, codes.OK, nil
}

func Encrypt(ctx context.Context, text string) (string, codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	key := viper.GetString("encryption_key")
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Error(err)
		return "", codes.Internal, err
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return encodeBase64(ciphertext), codes.OK, nil
}

func Decrypt(ctx context.Context, text string) (string, codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)	
	key := viper.GetString("encryption_key")
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Error(err)
		return "", codes.Internal, err
	}
	ciphertext, code, err := decodeBase64(ctx, text)
	if err != nil {
		return "", code, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), codes.OK, nil
}
