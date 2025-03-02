package emailservice

import (
	"auth/internal/config"
	"auth/internal/domain"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

type emailService struct {
	Emails              config.Emails
	SMTPDialer          *gomail.Dialer
	UserEmailRepository userEmailRepository
}

type userEmailRepository interface {
	GetUserEmail(userID uuid.UUID) (string, error)
}

func NewEmailService(
	emails config.Emails,
	smtpConfig config.SMTPServerConfig,
	userEmailRepository userEmailRepository,
) *emailService {

	smtpDialer := gomail.NewDialer(
		smtpConfig.Host, smtpConfig.Port,
		smtpConfig.Username, smtpConfig.Password)

	return &emailService{
		Emails:              emails,
		SMTPDialer:          smtpDialer,
		UserEmailRepository: userEmailRepository,
	}
}

func (s *emailService) SendSupportEmailToUser(
	userID uuid.UUID, emailContent domain.EmailContent,
) error {
	to, err := s.UserEmailRepository.GetUserEmail(userID)
	if err != nil {
		return errors.Wrap(err, "get user email")
	}

	return s.SendEmails(
		s.Emails.SupportEmail, []string{to},
		emailContent.Subject, emailContent.ContentType,
		emailContent.Body)
}

func (s *emailService) SendEmails(
	from string, to []string,
	subject, contentType, body string,
) error {
	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", to...)
	message.SetHeader("Subject", subject)
	message.SetBody(contentType, body)
	return s.SMTPDialer.DialAndSend(message)
}
