package models

import (
	"testing"
	"time"
)

func TestFlashMessage_GetStatus(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)
	lastWeek := now.AddDate(0, 0, -7)

	tests := []struct {
		name        string
		flashMsg    FlashMessage
		wantStatus  string
	}{
		{
			name: "Inactive message should return Inactive",
			flashMsg: FlashMessage{
				Active:    false,
				StartDate: yesterday,
				EndDate:   tomorrow,
			},
			wantStatus: "Inactive",
		},
		{
			name: "Current message should return Active",
			flashMsg: FlashMessage{
				Active:    true,
				StartDate: yesterday,
				EndDate:   tomorrow,
			},
			wantStatus: "Active",
		},
		{
			name: "Future message should return Scheduled",
			flashMsg: FlashMessage{
				Active:    true,
				StartDate: tomorrow,
				EndDate:   nextWeek,
			},
			wantStatus: "Scheduled",
		},
		{
			name: "Past message should return Expired",
			flashMsg: FlashMessage{
				Active:    true,
				StartDate: lastWeek,
				EndDate:   yesterday,
			},
			wantStatus: "Expired",
		},
		{
			name: "Today is start date should return Active",
			flashMsg: FlashMessage{
				Active:    true,
				StartDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
				EndDate:   tomorrow,
			},
			wantStatus: "Active",
		},
		{
			name: "Today is end date should return Active",
			flashMsg: FlashMessage{
				Active:    true,
				StartDate: yesterday,
				EndDate:   time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			},
			wantStatus: "Active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flashMsg.GetStatus(); got != tt.wantStatus {
				t.Errorf("FlashMessage.GetStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}