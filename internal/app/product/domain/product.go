package domain

import "time"

// ProductStatus represents the lifecycle state of a product.
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusActive    ProductStatus = "active"
	ProductStatusInactive  ProductStatus = "inactive"
	ProductStatusArchived  ProductStatus = "archived"
)

// Field names for change tracking.
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
	FieldArchivedAt  = "archived_at"
)

// Product is the aggregate root for product-related behavior.
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	archivedAt  *time.Time

	createdAt time.Time
	updatedAt time.Time

	changes *ChangeTracker
	events  []DomainEvent
}

// NewProduct constructs a new Product aggregate.
// All invariants are enforced here.
func NewProduct(
	id string,
	name string,
	description string,
	category string,
	basePrice *Money,
	now time.Time,
) *Product {
	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusInactive,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
	}

	p.changes.MarkDirty(FieldName)
	p.changes.MarkDirty(FieldDescription)
	p.changes.MarkDirty(FieldCategory)
	p.changes.MarkDirty(FieldBasePrice)
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, ProductCreatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})

	return p
}

// RehydrateProduct reconstructs a Product from persisted state.
// It does not emit events or mark fields as dirty.
func RehydrateProduct(
	id string,
	name string,
	description string,
	category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	archivedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		archivedAt:  archivedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		changes:     NewChangeTracker(),
	}
}

// ID returns product identifier.
func (p *Product) ID() string { return p.id }

func (p *Product) Name() string        { return p.name }
func (p *Product) Description() string { return p.description }
func (p *Product) Category() string    { return p.category }
func (p *Product) BasePrice() *Money   { return p.basePrice }
func (p *Product) Discount() *Discount { return p.discount }
func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) ArchivedAt() *time.Time { return p.archivedAt }
func (p *Product) CreatedAt() time.Time   { return p.createdAt }
func (p *Product) UpdatedAt() time.Time   { return p.updatedAt }

func (p *Product) Changes() *ChangeTracker { return p.changes }

// UpdateDetails updates name, description and category.
func (p *Product) UpdateDetails(name, description, category string, now time.Time) {
	changed := false

	if name != "" && name != p.name {
		p.name = name
		p.changes.MarkDirty(FieldName)
		changed = true
	}
	if description != "" && description != p.description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
		changed = true
	}
	if category != "" && category != p.category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
		changed = true
	}

	if changed {
		p.updatedAt = now
		p.events = append(p.events, ProductUpdatedEvent{
			baseEvent: baseEvent{occurredAt: now},
			ProductID: p.id,
		})
	}
}

// Activate switches product to active state.
func (p *Product) Activate(now time.Time) {
	if p.status == ProductStatusActive {
		return
	}
	if p.status == ProductStatusArchived {
		// archived products are immutable
		return
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, ProductActivatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
}

// Deactivate switches product to inactive state.
func (p *Product) Deactivate(now time.Time) {
	if p.status == ProductStatusInactive {
		return
	}
	if p.status == ProductStatusArchived {
		return
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, ProductDeactivatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
}

// Archive marks the product as archived (soft delete).
func (p *Product) Archive(now time.Time) {
	if p.status == ProductStatusArchived {
		return
	}

	p.status = ProductStatusArchived
	p.archivedAt = &now
	p.updatedAt = now

	p.changes.MarkDirty(FieldStatus)
	p.changes.MarkDirty(FieldArchivedAt)
}

// ApplyDiscount applies or replaces a discount.
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if discount == nil || !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, DiscountAppliedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})

	return nil
}

// RemoveDiscount clears current discount if any.
func (p *Product) RemoveDiscount(now time.Time) {
	if p.discount == nil {
		return
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, DiscountRemovedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
}

// DomainEvents returns a copy of pending events.
func (p *Product) DomainEvents() []DomainEvent {
	out := make([]DomainEvent, len(p.events))
	copy(out, p.events)
	return out
}

// ClearDomainEvents removes all pending events. Usually called after persistence.
func (p *Product) ClearDomainEvents() {
	p.events = nil
}

