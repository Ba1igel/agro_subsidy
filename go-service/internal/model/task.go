package model

import "time"

// SubsidiesTask is the raw event consumed from Kafka.
type SubsidiesTask struct {
	ID                     string    `json:"id"`
	FarmerID               string    `json:"farmer_id"`
	RegionCode             string    `json:"region_code"`
	RegionPriority         int       `json:"region_priority"`
	AmountRequested        float64   `json:"amount_requested"`
	AmountNorm             float64   `json:"amount_norm"`
	ApplicationDate        time.Time `json:"application_date"`
	CropType               string    `json:"crop_type"`
	FarmSizeHa             float64   `json:"farm_size_ha"`
	PreviousSubsidiesCount int       `json:"previous_subsidies_count"`
}

// MLRequest is the enriched payload sent to FastAPI.
// Derived fields (AmountRatio, Month, DayOfYear) are computed here so
// the ML service stays stateless and knows nothing about Kafka.
type MLRequest struct {
	TaskID                 string  `json:"task_id"`
	RegionPriority         int     `json:"region_priority"`
	AmountRequested        float64 `json:"amount_requested"`
	AmountNorm             float64 `json:"amount_norm"`
	AmountRatio            float64 `json:"amount_ratio"`
	Month                  int     `json:"month"`
	DayOfYear              int     `json:"day_of_year"`
	CropType               string  `json:"crop_type"`
	FarmSizeHa             float64 `json:"farm_size_ha"`
	PreviousSubsidiesCount int     `json:"previous_subsidies_count"`
}

// MLResponse is the scored result returned by FastAPI.
type MLResponse struct {
	TaskID     string             `json:"task_id"`
	Score      float64            `json:"score"`
	SHAPValues map[string]float64 `json:"shap_values"`
	Flags      []string           `json:"flags"`
}

// ToMLRequest computes derived features and builds the ML payload.
func (t *SubsidiesTask) ToMLRequest() MLRequest {
	ratio := 0.0
	if t.AmountNorm > 0 {
		ratio = t.AmountRequested / t.AmountNorm
	}
	return MLRequest{
		TaskID:                 t.ID,
		RegionPriority:         t.RegionPriority,
		AmountRequested:        t.AmountRequested,
		AmountNorm:             t.AmountNorm,
		AmountRatio:            ratio,
		Month:                  int(t.ApplicationDate.Month()),
		DayOfYear:              t.ApplicationDate.YearDay(),
		CropType:               t.CropType,
		FarmSizeHa:             t.FarmSizeHa,
		PreviousSubsidiesCount: t.PreviousSubsidiesCount,
	}
}
