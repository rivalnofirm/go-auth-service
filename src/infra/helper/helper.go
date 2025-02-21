package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"go-auth-service/src/infra/constants/common"
	errorMessage "go-auth-service/src/infra/constants/error_message"
	"go-auth-service/src/infra/models"
	"golang.org/x/crypto/bcrypt"
	"io"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var jwtKey = []byte(common.JwtKey)
var jwtRefreshKey = []byte(common.JwtRefreshKey)
var key = []byte(common.EncryptKey)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword, inputPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
}

// TokenClaims menyimpan klaim JWT untuk akses token
type TokenClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// RefreshTokenClaims menyimpan klaim JWT untuk refresh token
type RefreshTokenClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// GenerateToken membuat token JWT
func GenerateToken(data *models.User) (string, error) {
	expirationTime := time.Now().Add(common.AccessTokenExp)
	claims := &TokenClaims{
		UserID: data.Id,
		Email:  data.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// GenerateRefreshToken membuat refresh token JWT
func GenerateRefreshToken(data *models.User) (string, error) {
	expirationTime := time.Now().Add(common.RefreshTokenExp)
	claims := &RefreshTokenClaims{
		UserID: data.Id,
		Email:  data.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtRefreshKey)
}

// VerifyToken memverifikasi token JWT
func VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf(errorMessage.ExpiredToken)
			}
		}
		return nil, fmt.Errorf(errorMessage.InvalidToken)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf(errorMessage.InvalidToken)
	}

	return claims, nil
}

// VerifyRefreshToken memverifikasi refresh token JWT
func VerifyRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtRefreshKey, nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf(errorMessage.ExpiredToken)
			}
		}
		return nil, fmt.Errorf(errorMessage.InvalidToken)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf(errorMessage.InvalidToken)
	}

	return claims, nil
}

func DateToStringByFormat(nullTime sql.NullTime, format string) string {
	if nullTime.Valid {
		if format == "" {
			format = "2006-01-02 15:04:05" // default format
		}
		return nullTime.Time.Format(format)
	}
	return ""
}

func secureFilename(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	extension := filepath.Ext(filename)

	name = strings.ReplaceAll(name, " ", "_")

	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, c := range invalidChars {
		name = strings.ReplaceAll(name, c, "")
	}

	return name + extension
}

func UploadPicture(fileHeader *multipart.FileHeader, userId int64) (string, error) {
	allowedType := map[string]bool{"png": true, "jpg": true, "jpeg": true}
	uploadPath := os.Getenv("PATH_UPLOAD")

	// Create directory if it doesn't exist
	err := os.MkdirAll(uploadPath, 0755)
	if err != nil {
		return "", err
	}

	extension := filepath.Ext(fileHeader.Filename)
	filetype := strings.ToLower(strings.TrimPrefix(extension, "."))
	if !allowedType[filetype] {
		return "", fmt.Errorf("file type not allowed")
	}

	if fileHeader.Size > common.AttachmentSizeLimit {
		return "", fmt.Errorf("file size exceeds limit")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Reset the file pointer to the beginning
	file.Seek(0, io.SeekStart)

	filename := fmt.Sprintf("/%v/%s", userId, secureFilename(fileHeader.Filename))

	uploadFilePath := filepath.Join(uploadPath, filename)

	if err = os.MkdirAll(filepath.Dir(uploadFilePath), os.ModePerm); err != nil {
		return "", err
	}

	f, err := os.Create(uploadFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Reset the file pointer to the beginning
	file.Seek(0, io.SeekStart)

	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func SendMail(to, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	from := "no-reply@gmail.com"
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		from,
		to,
		subject,
		body,
	)

	auth := smtp.PlainAuth("", username, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, errors.New("input to unpad is empty")
	}
	padding := src[len(src)-1]
	if int(padding) > aes.BlockSize || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	return src[:len(src)-int(padding)], nil
}

func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	paddedText := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(paddedText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedText)

	encoded := base64.RawURLEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

func Decrypt(encodedText string) (string, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(encodedText)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(ciphertext, ciphertext)

	unpaddedText, err := unpad(ciphertext)
	if err != nil {
		return "", err
	}

	return string(unpaddedText), nil
}

func GetRealIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func HashRefreshToken(refreshToken string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(refreshToken))
	if err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

func NormalizeUserAgent(userAgent string) string {
	// Tentukan OS
	device := "unknown"
	if strings.Contains(userAgent, "Windows") {
		device = "Windows"
	} else if strings.Contains(userAgent, "Macintosh") {
		device = "MacOS"
	} else if strings.Contains(userAgent, "Linux") {
		device = "Linux"
	} else if strings.Contains(userAgent, "Android") {
		device = "Android"
	} else if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		device = "iOS"
	}

	// Tentukan Browser
	browser := "unknown"
	if strings.Contains(userAgent, "Chrome") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		browser = "Safari"
	} else if strings.Contains(userAgent, "Edge") {
		browser = "Edge"
	} else if strings.Contains(userAgent, "Opera") || strings.Contains(userAgent, "OPR") {
		browser = "Opera"
	}

	// Gabungkan dan format key
	deviceKey := strings.ToLower(device + "_" + browser)

	// Encode agar aman di Redis
	return deviceKey
}
