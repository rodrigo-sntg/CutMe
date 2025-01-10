package infrastructure

import (
	"CutMe/domain"
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	FromEmail    string
	FromPassword string
	ProjectName  string
}

type emailNotifier struct {
	config EmailConfig
}

func NewEmailNotifier(config EmailConfig) domain.Notifier {
	return &emailNotifier{config: config}
}

func (n *emailNotifier) SendSuccessEmailWithLinks(to, uploadID, originalFileURL, processedFileURL string) error {
	subject := fmt.Sprintf("[%s] Processamento Concluído com Sucesso!", n.config.ProjectName)
	body := fmt.Sprintf(
		"Olá,\n\nSeu arquivo com ID %s foi processado com sucesso!\n\n"+
			"Aqui estão os links para download:\n"+
			"Arquivo Original: %s\n"+
			"Arquivo Processado: %s\n\n"+
			"Obrigado,\nEquipe %s",
		uploadID, originalFileURL, processedFileURL, n.config.ProjectName,
	)
	return n.sendEmail(to, subject, body)
}

func (n *emailNotifier) SendFailureEmail(to, uploadID, errorMsg string) error {
	subject := fmt.Sprintf("[%s] Falha no Processamento do Arquivo", n.config.ProjectName)
	body := fmt.Sprintf(
		"Olá,\n\nO processamento do arquivo com ID %s falhou devido ao seguinte erro:\n\n%s\n\n"+
			"Por favor, tente novamente.\n\nEquipe %s",
		uploadID, errorMsg, n.config.ProjectName,
	)
	return n.sendEmail(to, subject, body)
}

func (n *emailNotifier) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.config.FromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	dialer := gomail.NewDialer(n.config.SMTPHost, n.config.SMTPPort, n.config.FromEmail, n.config.FromPassword)
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("erro ao enviar e-mail para %s: %w", to, err)
	}
	return nil
}
