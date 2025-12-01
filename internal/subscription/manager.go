package subscription

import "github.com/onurceri/botla-co/pkg/logger"

var log = logger.New("INFO")

func UpgradeSubscription(userID, planType string) error {
    log.Info("subscription_upgrade_disabled", map[string]any{"user": userID, "plan": planType})
    return nil
}

func CheckExpiredSubscriptions() error {
    log.Info("subscription_check_expired_disabled", map[string]any{})
    return nil
}

func ProcessRenewal(userID string) error {
    log.Info("subscription_process_renewal_disabled", map[string]any{"user": userID})
    return nil
}

func SuspendSubscription(userID string) error {
    log.Info("subscription_suspend", map[string]any{"user": userID})
    return nil
}
