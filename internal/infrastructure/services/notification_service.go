// internal/infrastructure/services/notification_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"myapp/internal/application/interfaces"
	"myapp/internal/domain"
)

type notificationService struct {
	slackWebhookURL string
	emailServiceURL string
}

func NewNotificationService(slackWebhookURL, emailServiceURL string) interfaces.NotificationService {
	return &notificationService{
		slackWebhookURL: slackWebhookURL,
		emailServiceURL: emailServiceURL,
	}
}

func (s *notificationService) SendStockAlert(ctx context.Context, event domain.StockLimitAlertEvent) error {
	// Send to Slack
	slackMessage := fmt.Sprintf(
		"🚨 Stock alert for %s: %d/%d (%.0f%% full)",
		event.ProductName, event.Current.Value(), event.MaxLimit.Value(),
		event.Utilization,
	)
	
	log.Printf("Sending to Slack: %s", slackMessage)
	// Actual HTTP call to Slack would go here
	
	// Send email if critical (>90%)
	if event.Utilization > 90 {
		emailBody := fmt.Sprintf(
			"CRITICAL: Product %s is at %.0f%% capacity (%d/%d)",
			event.ProductName, event.Utilization,
			event.Current.Value(), event.MaxLimit.Value(),
		)
		log.Printf("Sending email: %s", emailBody)
		// Actual email sending logic
	}
	
	return nil
}

func (s *notificationService) SendLowStockAlert(ctx context.Context, product *domain.Product, threshold int) error {
	message := fmt.Sprintf(
		"⚠️ Low stock alert: %s has only %d units left (threshold: %d)",
		product.Name, product.CurrentStock.Value(), threshold,
	)
	
	log.Printf("Sending low stock alert: %s", message)
	// Actual notification logic
	return nil
}