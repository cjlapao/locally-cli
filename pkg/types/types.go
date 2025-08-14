package types

type RecordStatus string

const (
	RecordStatusActive    RecordStatus = "active"
	RecordStatusInactive  RecordStatus = "inactive"
	RecordStatusDeleted   RecordStatus = "deleted"
	RecordStatusPending   RecordStatus = "pending"
	RecordStatusArchived  RecordStatus = "archived"
	RecordStatusSuspended RecordStatus = "suspended"
	RecordStatusExpired   RecordStatus = "expired"
	RecordStatusRevoked   RecordStatus = "revoked"
	RecordStatusCancelled RecordStatus = "cancelled"
)
