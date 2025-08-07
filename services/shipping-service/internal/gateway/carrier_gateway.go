package gateway

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
)

// CarrierGateway defines the interface for carrier API integration
type CarrierGateway interface {
	// Rate calculation
	GetShippingRates(ctx context.Context, request *models.ShippingQuoteRequest) ([]*models.ShippingQuote, error)
	
	// Shipment creation
	CreateShipment(ctx context.Context, shipment *models.Shipment) (*CarrierShipmentResponse, error)
	
	// Label generation
	GenerateLabel(ctx context.Context, shipmentID string) (*LabelResponse, error)
	
	// Tracking
	TrackShipment(ctx context.Context, trackingNumber string) (*TrackingResponse, error)
	
	// Validation
	ValidateAddress(ctx context.Context, address *models.Address) (*AddressValidationResponse, error)
}

// CarrierShipmentResponse represents the response from creating a shipment
type CarrierShipmentResponse struct {
	CarrierTrackingID     string                 `json:"carrier_tracking_id"`
	TrackingNumber        string                 `json:"tracking_number"`
	LabelURL              string                 `json:"label_url"`
	EstimatedDeliveryDate *time.Time             `json:"estimated_delivery_date"`
	Cost                  decimal.Decimal        `json:"cost"`
	RawResponse           map[string]interface{} `json:"raw_response"`
}

// LabelResponse represents a shipping label response
type LabelResponse struct {
	LabelURL    string `json:"label_url"`
	LabelFormat string `json:"label_format"` // PDF, PNG, etc.
	LabelData   []byte `json:"label_data"`
}

// TrackingResponse represents tracking information from carrier
type TrackingResponse struct {
	TrackingNumber    string                          `json:"tracking_number"`
	Status            string                          `json:"status"`
	EstimatedDelivery *time.Time                      `json:"estimated_delivery"`
	ActualDelivery    *time.Time                      `json:"actual_delivery"`
	Events            []*models.ShipmentTrackingEvent `json:"events"`
	RawResponse       map[string]interface{}          `json:"raw_response"`
}

// AddressValidationResponse represents address validation result
type AddressValidationResponse struct {
	IsValid         bool            `json:"is_valid"`
	ValidatedAddress *models.Address `json:"validated_address"`
	Suggestions     []*models.Address `json:"suggestions"`
	Errors          []string        `json:"errors"`
}

// MockCarrierGateway implements CarrierGateway for testing and development
type MockCarrierGateway struct {
	carrier models.CarrierType
}

// NewMockCarrierGateway creates a new mock carrier gateway
func NewMockCarrierGateway(carrier models.CarrierType) *MockCarrierGateway {
	return &MockCarrierGateway{carrier: carrier}
}

// GetShippingRates returns mock shipping rates
func (m *MockCarrierGateway) GetShippingRates(ctx context.Context, request *models.ShippingQuoteRequest) ([]*models.ShippingQuote, error) {
	// Mock rates based on carrier and weight
	baseRate := decimal.NewFromFloat(10.00)
	
	switch m.carrier {
	case models.CarrierFedEx:
		baseRate = decimal.NewFromFloat(12.00)
	case models.CarrierUPS:
		baseRate = decimal.NewFromFloat(11.00)
	case models.CarrierUSPS:
		baseRate = decimal.NewFromFloat(8.00)
	}
	
	// Add weight-based pricing
	weightCost := request.WeightKg.Mul(decimal.NewFromFloat(2.50))
	totalCost := baseRate.Add(weightCost)
	
	quotes := []*models.ShippingQuote{
		{
			ShippingMethodID:      "mock-ground-" + string(m.carrier),
			ShippingMethodName:    string(m.carrier) + " Ground",
			CarrierName:           m.carrier,
			ServiceType:           models.ServiceGround,
			Cost:                  totalCost,
			EstimatedDeliveryDays: 5,
			EstimatedDeliveryDate: time.Now().AddDate(0, 0, 5),
			IsFreeShipping:        request.OrderValue.GreaterThan(decimal.NewFromFloat(75.00)),
		},
		{
			ShippingMethodID:      "mock-express-" + string(m.carrier),
			ShippingMethodName:    string(m.carrier) + " Express",
			CarrierName:           m.carrier,
			ServiceType:           models.ServiceExpress,
			Cost:                  totalCost.Mul(decimal.NewFromFloat(1.8)),
			EstimatedDeliveryDays: 2,
			EstimatedDeliveryDate: time.Now().AddDate(0, 0, 2),
			IsFreeShipping:        false,
		},
	}
	
	return quotes, nil
}

// CreateShipment creates a mock shipment
func (m *MockCarrierGateway) CreateShipment(ctx context.Context, shipment *models.Shipment) (*CarrierShipmentResponse, error) {
	trackingNumber := generateMockTrackingNumber(m.carrier)
	estimatedDelivery := time.Now().AddDate(0, 0, 5)
	
	response := &CarrierShipmentResponse{
		CarrierTrackingID:     "MOCK_" + trackingNumber,
		TrackingNumber:        trackingNumber,
		LabelURL:              "https://mock-carrier.com/labels/" + trackingNumber + ".pdf",
		EstimatedDeliveryDate: &estimatedDelivery,
		Cost:                  shipment.ShippingCost,
		RawResponse: map[string]interface{}{
			"mock": true,
			"carrier": m.carrier,
			"created_at": time.Now(),
		},
	}
	
	return response, nil
}

// GenerateLabel generates a mock label
func (m *MockCarrierGateway) GenerateLabel(ctx context.Context, shipmentID string) (*LabelResponse, error) {
	response := &LabelResponse{
		LabelURL:    "https://mock-carrier.com/labels/" + shipmentID + ".pdf",
		LabelFormat: "PDF",
		LabelData:   []byte("mock label data"),
	}
	
	return response, nil
}

// TrackShipment returns mock tracking information
func (m *MockCarrierGateway) TrackShipment(ctx context.Context, trackingNumber string) (*TrackingResponse, error) {
	now := time.Now()
	events := []*models.ShipmentTrackingEvent{
		{
			Status:      "picked_up",
			Description: "Package picked up",
			Location:    "Origin Facility",
			EventTime:   now.AddDate(0, 0, -2),
		},
		{
			Status:      "in_transit",
			Description: "Package in transit",
			Location:    "Sorting Facility",
			EventTime:   now.AddDate(0, 0, -1),
		},
		{
			Status:      "out_for_delivery",
			Description: "Out for delivery",
			Location:    "Local Facility",
			EventTime:   now,
		},
	}
	
	response := &TrackingResponse{
		TrackingNumber:    trackingNumber,
		Status:            "out_for_delivery",
		EstimatedDelivery: &[]time.Time{now.AddDate(0, 0, 1)}[0],
		Events:            events,
		RawResponse: map[string]interface{}{
			"mock": true,
			"carrier": m.carrier,
		},
	}
	
	return response, nil
}

// ValidateAddress validates an address (mock implementation)
func (m *MockCarrierGateway) ValidateAddress(ctx context.Context, address *models.Address) (*AddressValidationResponse, error) {
	// Simple mock validation - just check if required fields are present
	isValid := address.Street != "" && address.City != "" && 
			   address.State != "" && address.PostalCode != "" && address.Country != ""
	
	response := &AddressValidationResponse{
		IsValid:          isValid,
		ValidatedAddress: address,
		Suggestions:      []*models.Address{},
		Errors:           []string{},
	}
	
	if !isValid {
		response.Errors = append(response.Errors, "Missing required address fields")
	}
	
	return response, nil
}

// generateMockTrackingNumber generates a mock tracking number
func generateMockTrackingNumber(carrier models.CarrierType) string {
	timestamp := time.Now().Unix()
	
	switch carrier {
	case models.CarrierFedEx:
		return "1Z" + string(rune(timestamp%1000000))
	case models.CarrierUPS:
		return "1Z" + string(rune(timestamp%1000000))
	case models.CarrierUSPS:
		return "9400" + string(rune(timestamp%100000000))
	default:
		return "MOCK" + string(rune(timestamp%1000000))
	}
}

// CarrierGatewayFactory creates carrier gateways
type CarrierGatewayFactory struct {
	gateways map[models.CarrierType]CarrierGateway
}

// NewCarrierGatewayFactory creates a new carrier gateway factory
func NewCarrierGatewayFactory() *CarrierGatewayFactory {
	return &CarrierGatewayFactory{
		gateways: make(map[models.CarrierType]CarrierGateway),
	}
}

// RegisterGateway registers a carrier gateway
func (f *CarrierGatewayFactory) RegisterGateway(carrier models.CarrierType, gateway CarrierGateway) {
	f.gateways[carrier] = gateway
}

// GetGateway returns a carrier gateway
func (f *CarrierGatewayFactory) GetGateway(carrier models.CarrierType) (CarrierGateway, bool) {
	gateway, exists := f.gateways[carrier]
	return gateway, exists
}

// GetAllGateways returns all registered gateways
func (f *CarrierGatewayFactory) GetAllGateways() map[models.CarrierType]CarrierGateway {
	return f.gateways
}
