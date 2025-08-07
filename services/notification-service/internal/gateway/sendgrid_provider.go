package gateway

import (
	"context"
	"fmt"
	"net/mail"
	"os"

	"github.com/sendgrid/sendgrid-go"
	sgMail "github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/shopsphere/shared/utils"
)

// SendGridProvider implements EmailProvider using SendGrid
type SendGridProvider struct {
	client   *sendgrid.Client
	fromName string
	fromEmail string
	logger   *utils.StructuredLogger
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(logger *utils.StructuredLogger) (*SendGridProvider, error) {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SENDGRID_API_KEY environment variable is required")
	}

	fromName := os.Getenv("SENDGRID_FROM_NAME")
	if fromName == "" {
		fromName = "ShopSphere"
	}

	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@shopsphere.com"
	}

	client := sendgrid.NewSendClient(apiKey)

	return &SendGridProvider{
		client:    client,
		fromName:  fromName,
		fromEmail: fromEmail,
		logger:    logger,
	}, nil
}

// SendEmail sends an email using SendGrid
func (sg *SendGridProvider) SendEmail(ctx context.Context, to, subject, body string) (*ProviderResponse, error) {
	from := sgMail.NewEmail(sg.fromName, sg.fromEmail)
	toEmail := sgMail.NewEmail("", to)
	
	message := sgMail.NewSingleEmail(from, subject, toEmail, body, body)

	response, err := sg.client.SendWithContext(ctx, message)
	if err != nil {
		sg.logger.Error(ctx, "Failed to send email via SendGrid", err, map[string]interface{}{
			"to":      to,
			"subject": subject,
		})
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	sg.logger.Info(ctx, "Email sent via SendGrid", map[string]interface{}{
		"to":           to,
		"subject":      subject,
		"status_code":  response.StatusCode,
		"message_id":   response.Headers["X-Message-Id"],
	})

	providerResponse := &ProviderResponse{
		Status: "sent",
		Metadata: map[string]interface{}{
			"status_code": response.StatusCode,
			"headers":     response.Headers,
		},
	}

	// Extract message ID from headers if available
	if messageIDs, ok := response.Headers["X-Message-Id"]; ok && len(messageIDs) > 0 {
		providerResponse.MessageID = messageIDs[0]
	}

	// Check for errors in response
	if response.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("SendGrid API error: status code %d", response.StatusCode)
		providerResponse.ErrorMessage = &errorMsg
		providerResponse.Status = "failed"
	}

	return providerResponse, nil
}

// SendTemplateEmail sends an email using a SendGrid template
func (sg *SendGridProvider) SendTemplateEmail(ctx context.Context, to, templateID string, variables map[string]interface{}) (*ProviderResponse, error) {
	from := sgMail.NewEmail(sg.fromName, sg.fromEmail)
	toEmail := sgMail.NewEmail("", to)
	
	message := sgMail.NewV3Mail()
	message.SetFrom(from)
	message.SetTemplateID(templateID)

	personalization := sgMail.NewPersonalization()
	personalization.AddTos(toEmail)
	
	// Add dynamic template data
	for key, value := range variables {
		personalization.SetDynamicTemplateData(key, value)
	}
	
	message.AddPersonalizations(personalization)

	response, err := sg.client.SendWithContext(ctx, message)
	if err != nil {
		sg.logger.Error(ctx, "Failed to send template email via SendGrid", err, map[string]interface{}{
			"to":          to,
			"template_id": templateID,
		})
		return nil, fmt.Errorf("failed to send template email: %w", err)
	}

	sg.logger.Info(ctx, "Template email sent via SendGrid", map[string]interface{}{
		"to":          to,
		"template_id": templateID,
		"status_code": response.StatusCode,
		"message_id":  response.Headers["X-Message-Id"],
	})

	providerResponse := &ProviderResponse{
		Status: "sent",
		Metadata: map[string]interface{}{
			"status_code": response.StatusCode,
			"template_id": templateID,
			"headers":     response.Headers,
		},
	}

	// Extract message ID from headers if available
	if messageIDs, ok := response.Headers["X-Message-Id"]; ok && len(messageIDs) > 0 {
		providerResponse.MessageID = messageIDs[0]
	}

	// Check for errors in response
	if response.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("SendGrid API error: status code %d", response.StatusCode)
		providerResponse.ErrorMessage = &errorMsg
		providerResponse.Status = "failed"
	}

	return providerResponse, nil
}

// ValidateEmail validates an email address
func (sg *SendGridProvider) ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}
	return nil
}
