package payment

import "fmt"

type MockPayer struct {
	txs      map[string]transaction
	nextID   int
	notified []string
}

type transaction struct {
	ID       string
	Amount   int64
	Refunded int64
	Currency string
	UserID   string
}

func NewMockPayer() *MockPayer {
	return &MockPayer{
		txs: make(map[string]transaction),
	}
}

func (m *MockPayer) Pay(userID string, amount int64, currency string) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("mock pay user=%s amount=%d: %w", userID, amount, ErrInvalidAmount)
	}

	m.nextID++
	txID := fmt.Sprintf("tx_%d", m.nextID)
	m.txs[txID] = transaction{
		ID:       txID,
		Amount:   amount,
		Currency: currency,
		UserID:   userID,
	}

	return txID, nil
}

func (m *MockPayer) Refund(txID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("mock refund tx=%s amount=%d: %w", txID, amount, ErrInvalidAmount)
	}

	tx, ok := m.txs[txID]
	if !ok {
		return fmt.Errorf("mock refund tx=%s: %w", txID, ErrTransactionNotFound)
	}

	remaining := tx.Amount - tx.Refunded
	if remaining == 0 {
		return fmt.Errorf("mock refund tx=%s: %w", txID, ErrAlreadyFullyRefunded)
	}
	if amount > remaining {
		return fmt.Errorf("mock refund tx=%s amount=%d exceeds remaining amount %d: %w", txID, amount, remaining, ErrInsufficientFunds)
	}
	tx.Refunded += amount
	m.txs[txID] = tx

	return nil
}

func (m *MockPayer) Notify(userID, message string) error {
	m.notified = append(m.notified, fmt.Sprintf("user=%s message=%s", userID, message))
	return nil
}
