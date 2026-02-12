package domain

// ChangeTracker tracks which fields of an aggregate have been modified.
// It is used by repositories to build targeted update mutations.
type ChangeTracker struct {
	dirtyFields map[string]bool
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

func (ct *ChangeTracker) MarkDirty(field string) {
	if ct == nil {
		return
	}
	ct.dirtyFields[field] = true
}

func (ct *ChangeTracker) Dirty(field string) bool {
	if ct == nil {
		return false
	}
	return ct.dirtyFields[field]
}

// Clear resets all dirty flags. Typically called after a successful commit.
func (ct *ChangeTracker) Clear() {
	if ct == nil {
		return
	}
	for k := range ct.dirtyFields {
		delete(ct.dirtyFields, k)
	}
}

