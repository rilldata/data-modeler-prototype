package auth

// Permission represents runtime access permissions.
type Permission int

const (
	// System-level permissions
	ManageInstances Permission = 0x01

	// Instance-level permissions
	ReadInstance  Permission = 0x11
	EditInstance  Permission = 0x12
	EditTrigger   Permission = 0x20
	ReadRepo      Permission = 0x13
	EditRepo      Permission = 0x14
	ReadObjects   Permission = 0x15
	ReadOLAP      Permission = 0x16
	ReadMetrics   Permission = 0x17
	ReadProfiling Permission = 0x18
	ReadAPI       Permission = 0x19
	ReadResolvers Permission = 0x1A
)
