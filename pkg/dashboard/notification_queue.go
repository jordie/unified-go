package dashboard

import (
	"fmt"
	"sync"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInApp   NotificationType = "in_app"
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypePush    NotificationType = "push"
	NotificationTypeSMS     NotificationType = "sms"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusSent       NotificationStatus = "sent"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusCancelled  NotificationStatus = "cancelled"
)

// Notification represents a notification to be sent
type Notification struct {
	ID               uint
	UserID           uint
	Username         string
	NotificationType NotificationType
	Status           NotificationStatus
	Subject          string
	Message          string
	Icon             string
	ActionURL        string
	CreatedAt        time.Time
	SentAt           *time.Time
	DeliveredAt      *time.Time
	FailedAt         *time.Time
	ExpiresAt        *time.Time
	Metadata         map[string]interface{}
	RetryCount       int
	MaxRetries       int
	LastError        string
	Priority         int // 0-10, higher = more urgent
}

// NotificationQueue manages notification delivery
type NotificationQueue struct {
	notifications    map[uint]*Notification // [notificationID]notification
	userQueue        map[uint][]*Notification // [userID][]notifications
	pendingQueue     []*Notification
	failedQueue      []*Notification
	maxQueueSize     int
	maxRetries       int
	retryBackoff     time.Duration
	deliveryHandlers map[NotificationType]DeliveryHandler
	mu               sync.RWMutex
	processingTicker *time.Ticker
	stopChan         chan bool
}

// DeliveryHandler is a function that delivers a notification
type DeliveryHandler func(*Notification) error

// NewNotificationQueue creates a new notification queue
func NewNotificationQueue() *NotificationQueue {
	nq := &NotificationQueue{
		notifications:    make(map[uint]*Notification),
		userQueue:        make(map[uint][]*Notification),
		pendingQueue:     make([]*Notification, 0),
		failedQueue:      make([]*Notification, 0),
		maxQueueSize:     10000,
		maxRetries:       3,
		retryBackoff:     5 * time.Second,
		deliveryHandlers: make(map[NotificationType]DeliveryHandler),
		stopChan:         make(chan bool),
	}

	// Start background processor
	go nq.processQueue()

	return nq
}

// QueueNotification adds a notification to the queue
func (nq *NotificationQueue) QueueNotification(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	nq.mu.Lock()
	defer nq.mu.Unlock()

	// Check queue size
	if len(nq.notifications) >= nq.maxQueueSize {
		return fmt.Errorf("notification queue full")
	}

	// Generate ID
	notification.ID = nq.generateNotificationID()
	notification.CreatedAt = time.Now()
	notification.Status = NotificationStatusPending
	notification.RetryCount = 0
	notification.MaxRetries = nq.maxRetries

	// Set expiration (24 hours by default)
	if notification.ExpiresAt == nil {
		expiration := time.Now().Add(24 * time.Hour)
		notification.ExpiresAt = &expiration
	}

	// Store notification
	nq.notifications[notification.ID] = notification
	nq.userQueue[notification.UserID] = append(nq.userQueue[notification.UserID], notification)
	nq.pendingQueue = append(nq.pendingQueue, notification)

	return nil
}

// RegisterDeliveryHandler registers a handler for a notification type
func (nq *NotificationQueue) RegisterDeliveryHandler(notificationType NotificationType, handler DeliveryHandler) {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	nq.deliveryHandlers[notificationType] = handler
}

// processQueue processes pending notifications in background
func (nq *NotificationQueue) processQueue() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nq.processPendingNotifications()
		case <-nq.stopChan:
			return
		}
	}
}

// processPendingNotifications processes notifications in the pending queue
func (nq *NotificationQueue) processPendingNotifications() {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	now := time.Now()
	remaining := make([]*Notification, 0)

	for _, notif := range nq.pendingQueue {
		// Check if expired
		if notif.ExpiresAt != nil && now.After(*notif.ExpiresAt) {
			notif.Status = NotificationStatusCancelled
			continue
		}

		// Try to deliver
		handler, exists := nq.deliveryHandlers[notif.NotificationType]
		if !exists {
			// No handler, mark as failed
			notif.Status = NotificationStatusFailed
			notif.LastError = "no delivery handler registered"
			now := time.Now()
			notif.FailedAt = &now
			nq.failedQueue = append(nq.failedQueue, notif)
			continue
		}

		err := handler(notif)
		if err == nil {
			// Success
			notif.Status = NotificationStatusSent
			now := time.Now()
			notif.SentAt = &now
		} else {
			// Failed, check retry count
			notif.LastError = err.Error()
			notif.RetryCount++

			if notif.RetryCount >= notif.MaxRetries {
				notif.Status = NotificationStatusFailed
				now := time.Now()
				notif.FailedAt = &now
				nq.failedQueue = append(nq.failedQueue, notif)
			} else {
				// Keep in pending for retry
				remaining = append(remaining, notif)
			}
		}
	}

	nq.pendingQueue = remaining
}

// MarkAsDelivered marks a notification as delivered
func (nq *NotificationQueue) MarkAsDelivered(notificationID uint) error {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	notif, exists := nq.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found")
	}

	notif.Status = NotificationStatusDelivered
	now := time.Now()
	notif.DeliveredAt = &now

	return nil
}

// GetUserNotifications returns notifications for a user
func (nq *NotificationQueue) GetUserNotifications(userID uint, limit int) []*Notification {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	notifications, exists := nq.userQueue[userID]
	if !exists {
		return make([]*Notification, 0)
	}

	if limit <= 0 || limit > len(notifications) {
		limit = len(notifications)
	}

	result := make([]*Notification, limit)
	copy(result, notifications[len(notifications)-limit:])

	return result
}

// GetPendingNotifications returns notifications awaiting delivery
func (nq *NotificationQueue) GetPendingNotifications() []*Notification {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	result := make([]*Notification, len(nq.pendingQueue))
	copy(result, nq.pendingQueue)

	return result
}

// GetFailedNotifications returns notifications that failed delivery
func (nq *NotificationQueue) GetFailedNotifications() []*Notification {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	result := make([]*Notification, len(nq.failedQueue))
	copy(result, nq.failedQueue)

	return result
}

// GetNotificationStatus returns the status of a notification
func (nq *NotificationQueue) GetNotificationStatus(notificationID uint) (NotificationStatus, error) {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	notif, exists := nq.notifications[notificationID]
	if !exists {
		return "", fmt.Errorf("notification not found")
	}

	return notif.Status, nil
}

// RetryFailedNotification retries a failed notification
func (nq *NotificationQueue) RetryFailedNotification(notificationID uint) error {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	notif, exists := nq.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found")
	}

	if notif.Status != NotificationStatusFailed {
		return fmt.Errorf("notification is not in failed status")
	}

	// Reset for retry
	notif.Status = NotificationStatusPending
	notif.RetryCount = 0
	notif.LastError = ""

	// Remove from failed queue
	remaining := make([]*Notification, 0)
	for _, n := range nq.failedQueue {
		if n.ID != notificationID {
			remaining = append(remaining, n)
		}
	}
	nq.failedQueue = remaining

	// Add back to pending queue
	nq.pendingQueue = append(nq.pendingQueue, notif)

	return nil
}

// CancelNotification cancels a notification
func (nq *NotificationQueue) CancelNotification(notificationID uint) error {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	notif, exists := nq.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found")
	}

	if notif.Status == NotificationStatusDelivered || notif.Status == NotificationStatusSent {
		return fmt.Errorf("cannot cancel notification that was already sent")
	}

	notif.Status = NotificationStatusCancelled

	return nil
}

// GetQueueStats returns statistics about the notification queue
func (nq *NotificationQueue) GetQueueStats() map[string]int {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	stats := make(map[string]int)
	stats["total"] = len(nq.notifications)
	stats["pending"] = len(nq.pendingQueue)
	stats["failed"] = len(nq.failedQueue)

	// Count by status
	statusCounts := make(map[NotificationStatus]int)
	for _, notif := range nq.notifications {
		statusCounts[notif.Status]++
	}

	stats["sent"] = statusCounts[NotificationStatusSent]
	stats["delivered"] = statusCounts[NotificationStatusDelivered]
	stats["cancelled"] = statusCounts[NotificationStatusCancelled]

	return stats
}

// PurgeOldNotifications removes old notifications from the queue
func (nq *NotificationQueue) PurgeOldNotifications(olderThan time.Duration) int {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	cutoffTime := time.Now().Add(-olderThan)
	purgedCount := 0

	// Create new maps and slices
	newNotifications := make(map[uint]*Notification)
	newUserQueue := make(map[uint][]*Notification)
	newPendingQueue := make([]*Notification, 0)
	newFailedQueue := make([]*Notification, 0)

	// Copy recent notifications
	for id, notif := range nq.notifications {
		if notif.CreatedAt.After(cutoffTime) {
			newNotifications[id] = notif

			// Add to user queue
			if len(newUserQueue[notif.UserID]) == 0 {
				newUserQueue[notif.UserID] = make([]*Notification, 0)
			}
			newUserQueue[notif.UserID] = append(newUserQueue[notif.UserID], notif)

			// Add to appropriate status queue
			if notif.Status == NotificationStatusPending {
				newPendingQueue = append(newPendingQueue, notif)
			} else if notif.Status == NotificationStatusFailed {
				newFailedQueue = append(newFailedQueue, notif)
			}
		} else {
			purgedCount++
		}
	}

	nq.notifications = newNotifications
	nq.userQueue = newUserQueue
	nq.pendingQueue = newPendingQueue
	nq.failedQueue = newFailedQueue

	return purgedCount
}

// Close stops the notification queue processor
func (nq *NotificationQueue) Close() error {
	close(nq.stopChan)
	return nil
}

// generateNotificationID generates a unique notification ID
func (nq *NotificationQueue) generateNotificationID() uint {
	maxID := uint(0)
	for id := range nq.notifications {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}
