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
		return "", fmt.Errorf("%v: error getting user: %v", errs.InvalidCredentials, err)
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
			return "", fmt.Errorf("%v: error getting otp: %v", errs.InvalidCredentials, err)
		}
		codeMatch := subtle.ConstantTimeCompare([]byte(password), []byte(otp.Code)) == 1
		if !codeMatch || time.Now().After(otp.ExpiresAt) {
			return "", fmt.Errorf("%v: otp doesnt match or expired: %v", errs.InvalidCredentials, err)
		}
		a.OTP.Delete(ctx, email)
	}
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		Issuer:    "Payments",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
	}
	tokenString, err := a.signWithOpenBao(claims)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (a *Auth) Restore(ctx context.Context, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.IncorrectEmail, err)
	}
	otp, err := a.GenerateOTP()
	if err != nil {
		return err
	}
	otp.Email = email
	_, err = a.OTP.Create(ctx, otp)
	if err != nil {
		return err
	}
	err = a.SendOTP(otp)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) GenerateOTP() (models.OTP, error) {
	otp := models.OTP{}
	b := make([]byte, 3)
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
	from := "payments.system@mail.ru"
	password := os.Getenv("SMTP_PASSWORD")
	fmt.Println("pass: ", password)
	smtpPort := "465"
	subject := "Пароль для входа в платежную систему"
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

func (a *Auth) signWithOpenBao(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	signingString, err := token.SigningString()
	if err != nil {
		return "", err
	}

	inputBase64 := base64.StdEncoding.EncodeToString([]byte(signingString))
	data := map[string]any{
		"input": inputBase64,
	}

	secret, err := a.SecretClient.Logical().Write("transit/sign/jwt_key", data)
	if err != nil {
		return "", err
	}

	// Vault возвращает строку вида "vault:v1:<base64_signature>"
	vaultSignature := secret.Data["signature"].(string)
	rawB64Sig := strings.TrimPrefix(vaultSignature, "vault:v1:")

	// 2. Декодируем стандартный Base64 от Vault
	sigBytes, err := base64.StdEncoding.DecodeString(rawB64Sig)
	if err != nil {
		return "", err
	}

	// Для Ed25519 sigBytes — это ВСЕГДА ровно 64 байта чистой подписи.
	// Просто переводим в URL-safe формат для JWT.
	return signingString + "." + base64.RawURLEncoding.EncodeToString(sigBytes), nil
}

func (a *Auth) RestorePassword(ctx context.Context, email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.IncorrectEmail, err)
	}
	otp, err := a.GenerateOTP()
	if err != nil {
		return err
	}
	otp.Email = email
	_, err = a.OTP.Create(ctx, otp)
	if err != nil {
		return err
	}
	err = a.SendOTP(otp)
	if err != nil {
		return err
	}
	return nil
}
