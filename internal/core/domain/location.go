package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Location represents a row in the "locations" table
type Location struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterLocationRequest struct {
	Name      string  `json:"name" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

type NearestLocation struct {
	Location
	Distance float64 `json:"distance"`
}

func (n *NearestLocation) MarshalJSON() ([]byte, error) {
	var distance = fmt.Sprintf("%.2f meters", n.Distance)

	if n.Distance >= 1000 {
		distance = fmt.Sprintf("%.2f kilometers", n.Distance/1000)
	}

	return json.Marshal(struct {
		Location
		Distance string `json:"distance"`
	}{
		Location: n.Location,
		Distance: distance,
	})
}
