package items

import "time"

type Item struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ArrivalDate *time.Time `json:"arrivalDate,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	ArrivalDate string `json:"arrivalDate,omitempty" validate:"omitempty,len=8"` // YYYYMMDD format
}

func (r *CreateItemRequest) ToItem() Item {
	return Item{
		Name: r.Name,
	}
}

type UpdateItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	ArrivalDate string `json:"arrivalDate,omitempty" validate:"omitempty,len=8"` // YYYYMMDD format
}
