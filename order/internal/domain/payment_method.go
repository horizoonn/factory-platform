package domain

type PaymentMethod string

const (
	PaymentMethodUnknown       PaymentMethod = ""
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSBP           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

func (m PaymentMethod) Valid() bool {
	switch m {
	case PaymentMethodCard,
		PaymentMethodSBP,
		PaymentMethodCreditCard,
		PaymentMethodInvestorMoney:
		return true
	default:
		return false
	}
}
