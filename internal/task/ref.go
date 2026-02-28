package task

// Ref types per spec §5.3.
const (
	RefParent    = "parent"
	RefBlockedBy = "blocked-by"
	RefRelatesTo = "relates-to"
)

// Ref represents a cross-reference to another task.
type Ref struct {
	Type string
	ID   string
}
