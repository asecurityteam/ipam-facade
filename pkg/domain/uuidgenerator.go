package domain

// UUIDGenerator generates sufficiently random unique identifiers
type UUIDGenerator interface {
	NewUUIDString() (string, error)
}
