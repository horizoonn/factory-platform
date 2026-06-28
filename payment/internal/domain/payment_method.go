package domain

type PaymentMethod int

const (
	PaymentMethodUnknown PaymentMethod = iota
	PaymentMethodCard
	PaymentMethodSBP
	PaymentMethodCreditCard
	PaymentMethodInvestorMoney
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
