package payment

import "errors"

type Payer interface {
	Pay(userID string, amount int64, currency string) (string, error)
}

type Refunder interface {
	Refund(txID string, amount int64) error
}

type Notifier interface {
	Notify(userID, message string) error
}

type PaymentProcessor interface {
	Payer
	Notifier
}

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrAlreadyFullyRefunded = errors.New("transaction already fully refunded")
)
