package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ShipmentStatus represents the status of a shipment
type ShipmentStatus string

const (
	ShipmentPending    ShipmentStatus = "pending"
	ShipmentProcessing ShipmentStatus = "processing"
	ShipmentPickedUp   ShipmentStatus = "picked_up"
	ShipmentInTransit  ShipmentStatus = "in_transit"
	ShipmentOutForDelivery ShipmentStatus = "out_for_delivery"
	ShipmentDelivered  ShipmentStatus = "delivered"
	ShipmentException  ShipmentStatus = "exception"
	ShipmentReturned   ShipmentStatus = "returned"
	ShipmentCancelled  ShipmentStatus = "cancelled"
)

// CarrierType represents supported shipping carriers
type CarrierType string

const (
	CarrierFedEx CarrierType = "FedEx"
	CarrierUPS   CarrierType = "UPS"
	CarrierUSPS  CarrierType = "USPS"
	CarrierDHL   CarrierType = "DHL"
)

// ServiceType represents shipping service types
type ServiceType string

const (
	ServiceGround      ServiceType = "GROUND"
	ServiceExpress     ServiceType = "EXPRESS"
	ServiceExpressSaver ServiceType = "EXPRESS_SAVER"
	ServiceOvernight   ServiceType = "OVERNIGHT"
	Service2DayAir     ServiceType = "2ND_DAY_AIR"
	ServiceNextDayAir  ServiceType = "NEXT_DAY_AIR"
	ServicePriority    ServiceType = "PRIORITY"
)

// Address represents a shipping address
type Address struct {
	Name         string `json:"name" db:"name"`
	Company      string `json:"company,omitempty" db:"company"`
	AddressLine1 string `json:"address_line_1" db:"address_line_1"`
	AddressLine2 string `json:"address_line_2,omitempty" db:"address_line_2"`
	City         string `json:"city" db:"city"`
	State        string `json:"state" db:"state"`
	PostalCode   string `json:"postal_code" db:"postal_code"`
	Country      string `json:"country" db:"country"`
	Phone        string `json:"phone,omitempty" db:"phone"`
	Email        string `json:"email,omitempty" db:"email"`
}

// PackageDimensions represents package dimensions
type PackageDimensions struct {
	Length decimal.Decimal `json:"length" db:"length"`
	Width  decimal.Decimal `json:"width" db:"width"`
	Height decimal.Decimal `json:"height" db:"height"`
	Unit   string          `json:"unit" db:"unit"` // "cm" or "in"
}

// ShippingMethod represents a shipping method
type ShippingMethod struct {
	ID              string          `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	CarrierName     CarrierType     `json:"carrier_name" db:"carrier_name"`
	ServiceType     ServiceType     `json:"service_type" db:"service_type"`
	DeliveryTimeMin int             `json:"delivery_time_min" db:"delivery_time_min"` // hours
	DeliveryTimeMax int             `json:"delivery_time_max" db:"delivery_time_max"` // hours
	BaseCost        decimal.Decimal `json:"base_cost" db:"base_cost"`
	CostPerKg       decimal.Decimal `json:"cost_per_kg" db:"cost_per_kg"`
	CostPerKm       decimal.Decimal `json:"cost_per_km" db:"cost_per_km"`
	MaxWeightKg     decimal.Decimal `json:"max_weight_kg" db:"max_weight_kg"`
	MaxDimensions   string          `json:"max_dimensions_cm" db:"max_dimensions_cm"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// ShippingZone represents a shipping zone
type ShippingZone struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Countries   []string  `json:"countries" db:"countries"`
	States      []string  `json:"states" db:"states"`
	PostalCodes []string  `json:"postal_codes" db:"postal_codes"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ShippingRate represents shipping rates for method/zone combinations
type ShippingRate struct {
	ID                     string          `json:"id" db:"id"`
	ShippingMethodID       string          `json:"shipping_method_id" db:"shipping_method_id"`
	ShippingZoneID         string          `json:"shipping_zone_id" db:"shipping_zone_id"`
	MinWeightKg            decimal.Decimal `json:"min_weight_kg" db:"min_weight_kg"`
	MaxWeightKg            decimal.Decimal `json:"max_weight_kg" db:"max_weight_kg"`
	MinOrderValue          decimal.Decimal `json:"min_order_value" db:"min_order_value"`
	MaxOrderValue          decimal.Decimal `json:"max_order_value" db:"max_order_value"`
	Rate                   decimal.Decimal `json:"rate" db:"rate"`
	FreeShippingThreshold  decimal.Decimal `json:"free_shipping_threshold" db:"free_shipping_threshold"`
	IsActive               bool            `json:"is_active" db:"is_active"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
}

// Shipment represents a shipment
type Shipment struct {
	ID                    string                 `json:"id" db:"id"`
	OrderID               string                 `json:"order_id" db:"order_id"`
	UserID                string                 `json:"user_id" db:"user_id"`
	ShippingMethodID      string                 `json:"shipping_method_id" db:"shipping_method_id"`
	TrackingNumber        string                 `json:"tracking_number" db:"tracking_number"`
	CarrierTrackingID     string                 `json:"carrier_tracking_id" db:"carrier_tracking_id"`
	Status                ShipmentStatus         `json:"status" db:"status"`
	FromAddress           Address                `json:"from_address" db:"from_address"`
	ToAddress             Address                `json:"to_address" db:"to_address"`
	WeightKg              decimal.Decimal        `json:"weight_kg" db:"weight_kg"`
	Dimensions            string                 `json:"dimensions_cm" db:"dimensions_cm"`
	DeclaredValue         decimal.Decimal        `json:"declared_value" db:"declared_value"`
	ShippingCost          decimal.Decimal        `json:"shipping_cost" db:"shipping_cost"`
	InsuranceCost         decimal.Decimal        `json:"insurance_cost" db:"insurance_cost"`
	TotalCost             decimal.Decimal        `json:"total_cost" db:"total_cost"`
	EstimatedDeliveryDate *time.Time             `json:"estimated_delivery_date" db:"estimated_delivery_date"`
	ActualPickupDate      *time.Time             `json:"actual_pickup_date" db:"actual_pickup_date"`
	ActualDeliveryDate    *time.Time             `json:"actual_delivery_date" db:"actual_delivery_date"`
	CarrierResponse       map[string]interface{} `json:"carrier_response" db:"carrier_response"`
	LabelURL              string                 `json:"label_url" db:"label_url"`
	Notes                 string                 `json:"notes" db:"notes"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// ShipmentTrackingEvent represents a tracking event
type ShipmentTrackingEvent struct {
	ID              string                 `json:"id" db:"id"`
	ShipmentID      string                 `json:"shipment_id" db:"shipment_id"`
	Status          string                 `json:"status" db:"status"`
	Description     string                 `json:"description" db:"description"`
	Location        string                 `json:"location" db:"location"`
	EventTime       time.Time              `json:"event_time" db:"event_time"`
	CarrierEventID  string                 `json:"carrier_event_id" db:"carrier_event_id"`
	CarrierRawData  map[string]interface{} `json:"carrier_raw_data" db:"carrier_raw_data"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// CarrierConfig represents carrier configuration
type CarrierConfig struct {
	ID                  string                 `json:"id" db:"id"`
	CarrierName         CarrierType            `json:"carrier_name" db:"carrier_name"`
	APIEndpoint         string                 `json:"api_endpoint" db:"api_endpoint"`
	APIKeyEncrypted     string                 `json:"api_key_encrypted" db:"api_key_encrypted"`
	APISecretEncrypted  string                 `json:"api_secret_encrypted" db:"api_secret_encrypted"`
	WebhookSecret       string                 `json:"webhook_secret" db:"webhook_secret"`
	IsActive            bool                   `json:"is_active" db:"is_active"`
	RateLimitPerMinute  int                    `json:"rate_limit_per_minute" db:"rate_limit_per_minute"`
	TimeoutSeconds      int                    `json:"timeout_seconds" db:"timeout_seconds"`
	ConfigData          map[string]interface{} `json:"config_data" db:"config_data"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// ShippingQuote represents a shipping quote
type ShippingQuote struct {
	ShippingMethodID      string          `json:"shipping_method_id"`
	ShippingMethodName    string          `json:"shipping_method_name"`
	CarrierName           CarrierType     `json:"carrier_name"`
	ServiceType           ServiceType     `json:"service_type"`
	Cost                  decimal.Decimal `json:"cost"`
	EstimatedDeliveryDays int             `json:"estimated_delivery_days"`
	EstimatedDeliveryDate time.Time       `json:"estimated_delivery_date"`
	IsFreeShipping        bool            `json:"is_free_shipping"`
}

// ShippingQuoteRequest represents a request for shipping quotes
type ShippingQuoteRequest struct {
	FromAddress   Address         `json:"from_address"`
	ToAddress     Address         `json:"to_address"`
	WeightKg      decimal.Decimal `json:"weight_kg"`
	Dimensions    string          `json:"dimensions_cm,omitempty"`
	DeclaredValue decimal.Decimal `json:"declared_value"`
	OrderValue    decimal.Decimal `json:"order_value"`
}

// TrackingInfo represents consolidated tracking information
type TrackingInfo struct {
	TrackingNumber     string                  `json:"tracking_number"`
	CarrierTrackingID  string                  `json:"carrier_tracking_id"`
	Status             ShipmentStatus          `json:"status"`
	EstimatedDelivery  *time.Time              `json:"estimated_delivery"`
	ActualDelivery     *time.Time              `json:"actual_delivery"`
	Events             []ShipmentTrackingEvent `json:"events"`
	LastUpdated        time.Time               `json:"last_updated"`
}

// NewShipment creates a new shipment
func NewShipment(orderID, userID, shippingMethodID string, fromAddr, toAddr Address, weightKg decimal.Decimal) *Shipment {
	return &Shipment{
		OrderID:          orderID,
		UserID:           userID,
		ShippingMethodID: shippingMethodID,
		Status:           ShipmentPending,
		FromAddress:      fromAddr,
		ToAddress:        toAddr,
		WeightKg:         weightKg,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// NewShippingMethod creates a new shipping method
func NewShippingMethod(name, description string, carrier CarrierType, serviceType ServiceType, deliveryTimeMin, deliveryTimeMax int, baseCost decimal.Decimal) *ShippingMethod {
	return &ShippingMethod{
		Name:            name,
		Description:     description,
		CarrierName:     carrier,
		ServiceType:     serviceType,
		DeliveryTimeMin: deliveryTimeMin,
		DeliveryTimeMax: deliveryTimeMax,
		BaseCost:        baseCost,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// NewShippingZone creates a new shipping zone
func NewShippingZone(name, description string, countries []string) *ShippingZone {
	return &ShippingZone{
		Name:        name,
		Description: description,
		Countries:   countries,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// IsValidStatus checks if the shipment status is valid
func (s ShipmentStatus) IsValid() bool {
	switch s {
	case ShipmentPending, ShipmentProcessing, ShipmentPickedUp, ShipmentInTransit,
		 ShipmentOutForDelivery, ShipmentDelivered, ShipmentException, ShipmentReturned, ShipmentCancelled:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if the shipment can transition to the given status
func (s ShipmentStatus) CanTransitionTo(newStatus ShipmentStatus) bool {
	transitions := map[ShipmentStatus][]ShipmentStatus{
		ShipmentPending:        {ShipmentProcessing, ShipmentCancelled},
		ShipmentProcessing:     {ShipmentPickedUp, ShipmentCancelled, ShipmentException},
		ShipmentPickedUp:       {ShipmentInTransit, ShipmentException},
		ShipmentInTransit:      {ShipmentOutForDelivery, ShipmentException, ShipmentReturned},
		ShipmentOutForDelivery: {ShipmentDelivered, ShipmentException, ShipmentReturned},
		ShipmentException:      {ShipmentInTransit, ShipmentReturned, ShipmentCancelled},
		ShipmentDelivered:      {}, // Terminal state
		ShipmentReturned:       {}, // Terminal state
		ShipmentCancelled:      {}, // Terminal state
	}

	allowedTransitions, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}
