package constants

type AuthStatus int

const (
	Authorized = iota + 1
	Voided
	Refunded
	Captured
)

func (as AuthStatus) String() string {
	return [...]string{"Authorized", "Voided", "Refunded", "Captured"}[as-1]
}
