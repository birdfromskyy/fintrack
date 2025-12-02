package services

import (
	"fmt"
	"log"

	"auth-service/internal/config"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

func (s *EmailService) SendVerificationCode(to, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "auth-service - Код подтверждения")

	body := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
            <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                <h1 style="color: #333; text-align: center;">auth-service</h1>
                <h2 style="color: #555; text-align: center;">Подтверждение электронной почты</h2>
                <p style="color: #666; font-size: 16px; line-height: 1.5;">
                    Добро пожаловать в auth-service! Для завершения регистрации введите следующий код подтверждения:
                </p>
                <div style="background-color: #f8f9fa; border: 2px dashed #dee2e6; padding: 20px; margin: 20px 0; text-align: center;">
                    <span style="font-size: 32px; font-weight: bold; color: #007bff; letter-spacing: 5px;">%s</span>
                </div>
                <p style="color: #666; font-size: 14px; line-height: 1.5;">
                    Код действителен в течение 10 минут. Если вы не регистрировались в auth-service, проигнорируйте это письмо.
                </p>
                <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
                <p style="color: #999; font-size: 12px; text-align: center;">
                    © 2024 auth-service. Система учёта личных финансов.
                </p>
            </div>
        </body>
        </html>
    `, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.SMTPHost, s.config.SMTPPort, s.config.SMTPEmail, s.config.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Verification code sent to %s", to)
	return nil
}

func (s *EmailService) SendPasswordReset(to, resetLink string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "auth-service - Восстановление пароля")

	body := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
            <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                <h1 style="color: #333; text-align: center;">auth-service</h1>
                <h2 style="color: #555; text-align: center;">Восстановление пароля</h2>
                <p style="color: #666; font-size: 16px; line-height: 1.5;">
                    Вы запросили восстановление пароля для вашей учётной записи auth-service.
                </p>
                <p style="color: #666; font-size: 16px; line-height: 1.5;">
                    Нажмите на кнопку ниже, чтобы создать новый пароль:
                </p>
                <div style="text-align: center; margin: 30px 0;">
                    <a href="%s" style="background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-size: 16px;">
                        Восстановить пароль
                    </a>
                </div>
                <p style="color: #666; font-size: 14px; line-height: 1.5;">
                    Ссылка действительна в течение 1 часа. Если вы не запрашивали восстановление пароля, проигнорируйте это письмо.
                </p>
                <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
                <p style="color: #999; font-size: 12px; text-align: center;">
                    © 2024 auth-service. Система учёта личных финансов.
                </p>
            </div>
        </body>
        </html>
    `, resetLink)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.SMTPHost, s.config.SMTPPort, s.config.SMTPEmail, s.config.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send password reset email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Password reset email sent to %s", to)
	return nil
}
