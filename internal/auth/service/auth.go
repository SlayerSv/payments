package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/auth/repo"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/vault/api"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Auth         repo.Auth
	OTP          repo.OTP
	User         repo.User
	SecretClient *api.Client
}

func NewAuth(auth repo.Auth, otp repo.OTP, user repo.User, secretClient *api.Client) *Auth {
	return &Auth{auth, otp, user, secretClient}
}

func (a *Auth) Register(ctx context.Context, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.IncorrectEmail, err)
	}
	otp, err := a.GenerateOTP()
	if err != nil {
		return err
	}
	otp.Email = email
	_, err = a.Auth.Register(ctx, otp)
	if err != nil {
		return err
	}
	err = a.SendOTP(otp)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	user, err := a.User.GetByEmail(ctx, email)
	if err != nil {
		return "", errs.InvalidCredentials
	}
	loggedIn := false
	if user.PasswordHash != nil {
		err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
		if err == nil {
			loggedIn = true
		}
	}
	if !loggedIn {
		otp, err := a.OTP.Get(ctx, email)
		if err != nil {
			return "", errs.InvalidCredentials
		}
		codeMatch := subtle.ConstantTimeCompare([]byte(password), []byte(otp.Code)) == 1
		if !codeMatch || time.Now().After(otp.ExpiresAt) {
			return "", errs.InvalidCredentials
		}
		go a.OTP.Delete(ctx, email)
	}
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		Issuer:    "Payments",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
	}
	tokenString, err := a.signWithOpenBao(ctx, claims)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (a *Auth) GenerateOTP() (models.OTP, error) {
	otp := models.OTP{}
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return otp, err
	}
	pass := hex.EncodeToString(b)
	otp.Code = pass
	otp.ExpiresAt = time.Now().Add(time.Hour * 24)
	return otp, nil
}

func (a *Auth) SendOTP(otp models.OTP) error {
	smtpHost := "smtp.mail.ru"
	from := "smartbuy.store@mail.ru"
	password := os.Getenv("SMTP_PASSWORD")
	smtpPort := "465"
	subject := "Smartbuy temporary password"
	body := "Ваш одноразовый пароль для входа в систему: " + otp.Code +
		"\nДействителен до: " + otp.ExpiresAt.String() +
		"\nПосле входа в систему поменяйте пароль в личном кабинете"
	// 1. Формирование сообщения
	msg := []byte(
		"To: " + otp.Email + "\r\n" +
			"From: " + from + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			body + "\r\n")
	// 2. Аутентификация
	auth := smtp.PlainAuth("", from, password, smtpHost)
	// 3. Установка безопасного TLS соединения (Implicit TLS для порта 465)
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, &tls.Config{
		ServerName: smtpHost,
	})
	if err != nil {
		return fmt.Errorf("TLS Dial failed: %w", err)
	}
	defer conn.Close()
	// 4. Создание SMTP клиента
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("NewClient failed: %w", err)
	}
	defer client.Close()
	// 5. Аутентификация и отправка
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP Auth failed: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("mail command failed: %w", err)
	}
	if err = client.Rcpt(otp.Email); err != nil {
		return fmt.Errorf("rcpt command failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data command failed: %w", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}
	return client.Quit()
}

func (a *Auth) signWithOpenBao(ctx context.Context, claims jwt.Claims) (string, error) {
	// Создаем заголовок и полезную нагрузку
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Нам нужна часть токена без подписи (header.payload)
	signingString, err := token.SigningString()
	if err != nil {
		return "", err
	}
	// Отправляем в OpenBao на подпись
	// В OpenBao Transit это запрос: POST /v1/transit/sign/my-key-name
	// Payload: { "input": "base64_encoded_signing_string" }
	// Хранилище Transit ожидает данные в base64
	inputBase64 := base64.StdEncoding.EncodeToString([]byte(signingString))
	data := map[string]interface{}{
		"input": inputBase64,
	}
	// Путь: transit/sign/{имя_ключа}
	secret, err := a.SecretClient.Logical().Write("transit/sign/jwt-key", data)
	if err != nil {
		return "", err
	}
	// Bao вернет подпись в формате vault:v1:BASE64_SIGNATURE
	vaultSignature := secret.Data["signature"].(string)
	signature := strings.TrimPrefix(vaultSignature, "vault:v1:")
	// Собираем итоговый JWT: header.payload.signature
	return signingString + "." + signature, nil
}
