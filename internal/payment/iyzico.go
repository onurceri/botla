package payment

import "github.com/onurceri/botla-co/pkg/logger"

type User struct{ ID string }

type Payment struct {
    Amount int64
    PlanID string
    Token  string
}

var log = logger.New("INFO")

func CreateCustomer(user *User) (string, error) {
    log.Info("payment_create_customer_skipped", map[string]any{"user": user.ID})
    return "stub-customer-id", nil
}

func CreatePayment(payment *Payment) (string, error) {
    log.Info("payment_create_payment_skipped", map[string]any{"amount": payment.Amount, "plan": payment.PlanID})
    return "stub-transaction-id", nil
}

func GetPaymentStatus(transactionID string) (string, error) {
    log.Info("payment_get_status_stub", map[string]any{"transaction_id": transactionID})
    return "PENDING", nil
}

func CreateRecurringPayment(customerID string) error {
    log.Info("payment_create_recurring_skipped", map[string]any{"customer_id": customerID})
    return nil
}
