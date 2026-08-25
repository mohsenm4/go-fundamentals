package payment

import (
	"errors"
	"testing"
)

var (
	_ Payer            = (*MockPayer)(nil)
	_ Refunder         = (*MockPayer)(nil)
	_ Notifier         = (*MockPayer)(nil)
	_ PaymentProcessor = (*MockPayer)(nil)
)

func TestPay(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		amount   int64
		currency string
		wantErr  error
	}{
		{"valid payment", "user1", 100, "USD", nil},
		{"invalid amount", "user2", -50, "USD", ErrInvalidAmount},
		{"zero amount", "user3", 0, "USD", ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayer := NewMockPayer()
			txID, err := mockPayer.Pay(tt.userID, tt.amount, tt.currency)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Pay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && txID == "" {
				t.Errorf("Pay() returned empty transaction ID for valid payment")
			}
		})
	}
}

func TestRefund_StandardCases(t *testing.T) {
	tests := []struct {
		name         string
		refundAmount int64
		wantErr      error
	}{
		{"valid refund", 50, nil},
		{"invalid refund amount", -20, ErrInvalidAmount},
		{"refund zero amount", 0, ErrInvalidAmount},
		{"refund more than paid", 150, ErrInsufficientFunds},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayer := NewMockPayer()

			txID, err := mockPayer.Pay("user1", 100, "USD")
			if err != nil {
				t.Fatalf("Setup Pay() failed: %v", err)
			}

			err = mockPayer.Refund(txID, tt.refundAmount)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Refund() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefund_NonExistentTransaction(t *testing.T) {
	mockPayer := NewMockPayer()

	err := mockPayer.Refund("invalid-tx-id", 50)
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Errorf("Refund() error = %v, wantErr %v", err, ErrTransactionNotFound)
	}
}

func TestRefund_AlreadyFullyRefunded(t *testing.T) {
	mockPayer := NewMockPayer()

	txID, err := mockPayer.Pay("user1", 100, "USD")
	if err != nil {
		t.Fatalf("Setup Pay() failed: %v", err)
	}

	if err := mockPayer.Refund(txID, 100); err != nil {
		t.Fatalf("First Refund failed: %v", err)
	}

	err = mockPayer.Refund(txID, 50)
	if !errors.Is(err, ErrAlreadyFullyRefunded) {
		t.Errorf("Refund() error = %v, wantErr %v", err, ErrAlreadyFullyRefunded)
	}
}

func TestNotify(t *testing.T) {
	mockPayer := NewMockPayer()

	tests := []struct {
		name    string
		userID  string
		message string
		wantErr error
	}{
		{"valid notification", "user1", "Payment successful", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockPayer.Notify(tt.userID, tt.message)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Notify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
