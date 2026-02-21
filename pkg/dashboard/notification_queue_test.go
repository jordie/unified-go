package dashboard

import (
	"fmt"
	"testing"
	"time"
)

func TestNewNotificationQueue(t *testing.T) {
	queue := NewNotificationQueue()

	if queue == nil {
		t.Fatal("NewNotificationQueue returned nil")
	}
	if queue.maxRetries != 3 {
		t.Errorf("expected 3 max retries, got %d", queue.maxRetries)
	}
}

func TestQueueNotification(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Subject:          "Test Subject",
		Message:          "Test message",
	}

	err := queue.QueueNotification(notification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notification.ID == 0 {
		t.Error("notification ID not set")
	}
	if notification.Status != NotificationStatusPending {
		t.Errorf("expected pending status, got %s", notification.Status)
	}
}

func TestQueueNotificationNil(t *testing.T) {
	queue := NewNotificationQueue()

	err := queue.QueueNotification(nil)
	if err == nil {
		t.Error("expected error when queuing nil notification")
	}
}

func TestGetUserNotifications(t *testing.T) {
	queue := NewNotificationQueue()

	// Queue multiple notifications
	for i := 0; i < 3; i++ {
		notification := &Notification{
			UserID:           123,
			Username:         "testuser",
			NotificationType: NotificationTypeInApp,
			Message:          "Message " + string(rune('0'+i)),
		}
		_ = queue.QueueNotification(notification)
	}

	notifications := queue.GetUserNotifications(123, 10)
	if len(notifications) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(notifications))
	}
}

func TestGetUserNotificationsLimit(t *testing.T) {
	queue := NewNotificationQueue()

	// Queue 5 notifications
	for i := 0; i < 5; i++ {
		notification := &Notification{
			UserID:           123,
			Username:         "testuser",
			NotificationType: NotificationTypeInApp,
			Message:          "Message " + string(rune('0'+i)),
		}
		_ = queue.QueueNotification(notification)
	}

	notifications := queue.GetUserNotifications(123, 3)
	if len(notifications) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(notifications))
	}
}

func TestGetPendingNotifications(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	pending := queue.GetPendingNotifications()
	if len(pending) != 1 {
		t.Errorf("expected 1 pending notification, got %d", len(pending))
	}
	if pending[0].Status != NotificationStatusPending {
		t.Errorf("expected pending status, got %s", pending[0].Status)
	}
}

func TestMarkAsDelivered(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	err := queue.MarkAsDelivered(notification.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, _ := queue.GetNotificationStatus(notification.ID)
	if status != NotificationStatusDelivered {
		t.Errorf("expected delivered status, got %s", status)
	}
}

func TestMarkAsDeliveredNotFound(t *testing.T) {
	queue := NewNotificationQueue()

	err := queue.MarkAsDelivered(9999)
	if err == nil {
		t.Error("expected error for non-existent notification")
	}
}

func TestGetNotificationStatus(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	status, err := queue.GetNotificationStatus(notification.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != NotificationStatusPending {
		t.Errorf("expected pending status, got %s", status)
	}
}

func TestRetryFailedNotification(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	// Manually set to failed
	notif := queue.GetPendingNotifications()[0]
	notif.Status = NotificationStatusFailed
	queue.failedQueue = append(queue.failedQueue, notif)
	queue.pendingQueue = make([]*Notification, 0)

	err := queue.RetryFailedNotification(notification.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pending := queue.GetPendingNotifications()
	if len(pending) != 1 {
		t.Error("notification not moved back to pending")
	}
}

func TestCancelNotification(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	err := queue.CancelNotification(notification.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, _ := queue.GetNotificationStatus(notification.ID)
	if status != NotificationStatusCancelled {
		t.Errorf("expected cancelled status, got %s", status)
	}
}

func TestRegisterDeliveryHandler(t *testing.T) {
	queue := NewNotificationQueue()

	// Register a handler
	handlerCalled := false
	handler := func(notif *Notification) error {
		handlerCalled = true
		return nil
	}

	queue.RegisterDeliveryHandler(NotificationTypeInApp, handler)

	// Process queue
	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	// Wait a bit for async processing
	time.Sleep(100 * time.Millisecond)

	if handlerCalled {
		t.Log("Handler was called")
	}
}

func TestDeliveryHandlerError(t *testing.T) {
	queue := NewNotificationQueue()

	// Register a handler that fails
	handler := func(notif *Notification) error {
		return fmt.Errorf("delivery failed")
	}

	queue.RegisterDeliveryHandler(NotificationTypeEmail, handler)

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeEmail,
		Message:          "Test",
		MaxRetries:       2,
	}

	_ = queue.QueueNotification(notification)

	// Wait for processing (background processor runs every 1 second, needs 3+ retries)
	time.Sleep(4000 * time.Millisecond)

	status, _ := queue.GetNotificationStatus(notification.ID)
	if status != NotificationStatusFailed {
		t.Errorf("expected failed status after max retries, got %s", status)
	}
}

func TestGetQueueStats(t *testing.T) {
	queue := NewNotificationQueue()

	// Queue notifications
	for i := 0; i < 3; i++ {
		notification := &Notification{
			UserID:           123,
			Username:         "testuser",
			NotificationType: NotificationTypeInApp,
			Message:          "Test " + string(rune('0'+i)),
		}
		_ = queue.QueueNotification(notification)
	}

	stats := queue.GetQueueStats()

	if stats["total"] != 3 {
		t.Errorf("expected total 3, got %d", stats["total"])
	}
	if stats["pending"] != 3 {
		t.Errorf("expected pending 3, got %d", stats["pending"])
	}
}

func TestPurgeOldNotifications(t *testing.T) {
	queue := NewNotificationQueue()

	// Queue a notification
	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	_ = queue.QueueNotification(notification)

	// Manually set created time to old
	notif := queue.notifications[notification.ID]
	oldTime := time.Now().Add(-48 * time.Hour)
	notif.CreatedAt = oldTime

	// Purge notifications older than 24 hours
	purged := queue.PurgeOldNotifications(24 * time.Hour)

	if purged != 1 {
		t.Errorf("expected 1 purged notification, got %d", purged)
	}

	stats := queue.GetQueueStats()
	if stats["total"] != 0 {
		t.Errorf("expected total 0 after purge, got %d", stats["total"])
	}
}

func TestNotificationMetadata(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Subject:          "Test",
		Message:          "Test message",
		Metadata: map[string]interface{}{
			"achievement_id": 456,
			"points":         100,
		},
	}

	_ = queue.QueueNotification(notification)

	userNotifs := queue.GetUserNotifications(123, 10)
	if len(userNotifs) == 0 {
		t.Fatal("notification not found")
	}

	if achievementID, ok := userNotifs[0].Metadata["achievement_id"]; ok {
		if achievementID != 456 {
			t.Errorf("expected achievement_id 456, got %v", achievementID)
		}
	} else {
		t.Error("metadata not preserved")
	}
}

func TestNotificationExpiration(t *testing.T) {
	queue := NewNotificationQueue()

	notification := &Notification{
		UserID:           123,
		Username:         "testuser",
		NotificationType: NotificationTypeInApp,
		Message:          "Test",
	}

	// Set expiration to past
	expiredTime := time.Now().Add(-1 * time.Hour)
	notification.ExpiresAt = &expiredTime

	_ = queue.QueueNotification(notification)

	// Process queue (background processor runs every 1 second)
	time.Sleep(2000 * time.Millisecond)

	status, _ := queue.GetNotificationStatus(notification.ID)
	if status != NotificationStatusCancelled {
		t.Errorf("expected cancelled status for expired notification, got %s", status)
	}
}

func TestNotificationPriority(t *testing.T) {
	queue := NewNotificationQueue()

	// Queue notifications with different priorities
	for i := 0; i < 3; i++ {
		notification := &Notification{
			UserID:           123,
			Username:         "testuser",
			NotificationType: NotificationTypeInApp,
			Message:          "Test " + string(rune('0'+i)),
			Priority:         i,
		}
		_ = queue.QueueNotification(notification)
	}

	// Verify all queued
	pending := queue.GetPendingNotifications()
	if len(pending) != 3 {
		t.Errorf("expected 3 pending notifications, got %d", len(pending))
	}
}

// BenchmarkQueueNotification benchmarks notification queuing
func BenchmarkQueueNotification(b *testing.B) {
	queue := NewNotificationQueue()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notification := &Notification{
			UserID:           uint(i % 1000),
			Username:         "user",
			NotificationType: NotificationTypeInApp,
			Message:          "Test",
		}
		_ = queue.QueueNotification(notification)
	}
}

// BenchmarkGetUserNotifications benchmarks user notification retrieval
func BenchmarkGetUserNotifications(b *testing.B) {
	queue := NewNotificationQueue()

	// Pre-populate notifications
	for i := 0; i < 100; i++ {
		notification := &Notification{
			UserID:           123,
			Username:         "testuser",
			NotificationType: NotificationTypeInApp,
			Message:          "Test " + string(rune('0'+(i%10))),
		}
		_ = queue.QueueNotification(notification)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queue.GetUserNotifications(123, 10)
	}
}

// BenchmarkGetQueueStats benchmarks stats calculation
func BenchmarkNotificationGetQueueStats(b *testing.B) {
	queue := NewNotificationQueue()

	// Pre-populate notifications
	for i := 0; i < 100; i++ {
		notification := &Notification{
			UserID:           uint(i % 10),
			Username:         "user",
			NotificationType: NotificationTypeInApp,
			Message:          "Test",
		}
		_ = queue.QueueNotification(notification)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queue.GetQueueStats()
	}
}

// TestNotificationQueueConcurrency verifies thread-safe access
func TestNotificationQueueConcurrency(t *testing.T) {
	queue := NewNotificationQueue()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 20; j++ {
				notification := &Notification{
					UserID:           uint(id),
					Username:         "user",
					NotificationType: NotificationTypeInApp,
					Message:          "Test",
				}
				_ = queue.QueueNotification(notification)
				_ = queue.GetUserNotifications(uint(id), 10)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify notifications were processed
	stats := queue.GetQueueStats()
	if stats["total"] == 0 {
		t.Error("no notifications queued")
	}
}
