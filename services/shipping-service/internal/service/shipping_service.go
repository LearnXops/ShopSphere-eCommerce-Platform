package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/shipping-service/internal/gateway"
	"github.com/shopsphere/shipping-service/internal/repository"
)

// ShippingService handles shipping business logic
type ShippingService struct {
	repo           repository.ShippingRepository
	gatewayFactory *gateway.CarrierGatewayFactory
	logger         *utils.StructuredLogger
}

// NewShippingService creates a new shipping service
func NewShippingService(repo repository.ShippingRepository, gatewayFactory *gateway.CarrierGatewayFactory, logger *utils.StructuredLogger) *ShippingService {
	return &ShippingService{
		repo:           repo,
		gatewayFactory: gatewayFactory,
		logger:         logger,
	}
}

// GetShippingQuotes calculates shipping rates for a quote request
func (s *ShippingService) GetShippingQuotes(ctx context.Context, request *models.ShippingQuoteRequest) ([]*models.ShippingQuote, error) {
	s.logger.Info(ctx, "Getting shipping quotes", map[string]interface{}{
		"weight_kg": request.WeightKg,
		"country":   request.ToAddress.Country,
	})

	// Find appropriate shipping zone
	zone, err := s.repo.GetShippingZoneByAddress(ctx, &request.ToAddress)
	if err != nil {
		s.logger.Error(ctx, "Failed to find shipping zone", err, map[string]interface{}{
			"country": request.ToAddress.Country,
			"state":   request.ToAddress.State,
		})
		return nil, fmt.Errorf("failed to find shipping zone: %w", err)
	}

	// Get active shipping methods
	methods, err := s.repo.GetShippingMethods(ctx, true)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipping methods", err)
		return nil, fmt.Errorf("failed to get shipping methods: %w", err)
	}

	var quotes []*models.ShippingQuote

	// Calculate rates for each method
	for _, method := range methods {
		// Get rates for this method and zone
		rates, err := s.repo.GetShippingRates(ctx, method.ID, zone.ID)
		if err != nil {
			s.logger.Error(ctx, "Failed to get shipping rates", err, map[string]interface{}{
				"method_id": method.ID,
				"zone_id":   zone.ID,
			})
			continue
		}

		// Find applicable rate
		applicableRate := s.findApplicableRate(rates, request.WeightKg, request.OrderValue)
		if applicableRate == nil {
			continue
		}

		// Calculate cost
		cost := s.calculateShippingCost(method, applicableRate, request)
		
		// Check for free shipping
		isFreeShipping := request.OrderValue.GreaterThanOrEqual(applicableRate.FreeShippingThreshold)
		if isFreeShipping {
			cost = decimal.Zero
		}

		// Calculate estimated delivery
		estimatedDays := method.DeliveryTimeMin / 24 // Convert hours to days
		if estimatedDays == 0 {
			estimatedDays = 1
		}

		quote := &models.ShippingQuote{
			ShippingMethodID:      method.ID,
			ShippingMethodName:    method.Name,
			CarrierName:           method.CarrierName,
			ServiceType:           method.ServiceType,
			Cost:                  cost,
			EstimatedDeliveryDays: estimatedDays,
			EstimatedDeliveryDate: time.Now().AddDate(0, 0, estimatedDays),
			IsFreeShipping:        isFreeShipping,
		}

		quotes = append(quotes, quote)
	}

	s.logger.Info(ctx, "Generated shipping quotes", map[string]interface{}{
		"quote_count": len(quotes),
	})

	return quotes, nil
}

// CreateShipment creates a new shipment
func (s *ShippingService) CreateShipment(ctx context.Context, orderID, userID, shippingMethodID string, fromAddr, toAddr models.Address, weightKg decimal.Decimal, declaredValue decimal.Decimal) (*models.Shipment, error) {
	s.logger.Info(ctx, "Creating shipment", map[string]interface{}{
		"order_id":           orderID,
		"user_id":            userID,
		"shipping_method_id": shippingMethodID,
		"weight_kg":          weightKg,
	})

	// Get shipping method
	method, err := s.repo.GetShippingMethod(ctx, shippingMethodID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipping method", err, map[string]interface{}{
			"shipping_method_id": shippingMethodID,
		})
		return nil, fmt.Errorf("failed to get shipping method: %w", err)
	}

	// Calculate shipping cost
	zone, err := s.repo.GetShippingZoneByAddress(ctx, &toAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipping zone: %w", err)
	}

	rates, err := s.repo.GetShippingRates(ctx, method.ID, zone.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping rates: %w", err)
	}

	applicableRate := s.findApplicableRate(rates, weightKg, declaredValue)
	if applicableRate == nil {
		return nil, fmt.Errorf("no applicable shipping rate found")
	}

	shippingCost := s.calculateShippingCostFromRate(method, applicableRate, weightKg)

	// Create shipment
	shipment := models.NewShipment(orderID, userID, shippingMethodID, fromAddr, toAddr, weightKg)
	shipment.ID = uuid.New().String()
	shipment.DeclaredValue = declaredValue
	shipment.ShippingCost = shippingCost
	shipment.TotalCost = shippingCost // Add insurance cost if needed

	// Generate tracking number
	shipment.TrackingNumber = s.generateTrackingNumber(method.CarrierName)

	// Save to database
	err = s.repo.CreateShipment(ctx, shipment)
	if err != nil {
		s.logger.Error(ctx, "Failed to create shipment", err, map[string]interface{}{
			"shipment_id": shipment.ID,
		})
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Create shipment with carrier (async)
	go s.createCarrierShipment(context.Background(), shipment)

	s.logger.Info(ctx, "Shipment created successfully", map[string]interface{}{
		"shipment_id":     shipment.ID,
		"tracking_number": shipment.TrackingNumber,
	})

	return shipment, nil
}

// GetShipment retrieves a shipment by ID
func (s *ShippingService) GetShipment(ctx context.Context, id string) (*models.Shipment, error) {
	shipment, err := s.repo.GetShipment(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipment", err, map[string]interface{}{
			"shipment_id": id,
		})
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	return shipment, nil
}

// GetShipmentByTrackingNumber retrieves a shipment by tracking number
func (s *ShippingService) GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	shipment, err := s.repo.GetShipmentByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipment by tracking number", err, map[string]interface{}{
			"tracking_number": trackingNumber,
		})
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	return shipment, nil
}

// GetShipmentsByOrder retrieves shipments for an order
func (s *ShippingService) GetShipmentsByOrder(ctx context.Context, orderID string) ([]*models.Shipment, error) {
	shipments, err := s.repo.GetShipmentsByOrder(ctx, orderID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipments by order", err, map[string]interface{}{
			"order_id": orderID,
		})
		return nil, fmt.Errorf("failed to get shipments: %w", err)
	}

	return shipments, nil
}

// TrackShipment retrieves tracking information for a shipment
func (s *ShippingService) TrackShipment(ctx context.Context, trackingNumber string) (*models.TrackingInfo, error) {
	s.logger.Info(ctx, "Tracking shipment", map[string]interface{}{
		"tracking_number": trackingNumber,
	})

	// Get shipment from database
	shipment, err := s.repo.GetShipmentByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment: %w", err)
	}

	// Get tracking events from database
	events, err := s.repo.GetTrackingEvents(ctx, shipment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking events: %w", err)
	}

	// Get shipping method to determine carrier
	method, err := s.repo.GetShippingMethod(ctx, shipment.ShippingMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping method: %w", err)
	}

	// Try to get real-time tracking from carrier
	gateway, exists := s.gatewayFactory.GetGateway(method.CarrierName)
	if exists {
		carrierTracking, err := gateway.TrackShipment(ctx, trackingNumber)
		if err == nil && len(carrierTracking.Events) > 0 {
			// Update events with carrier data
			for _, carrierEvent := range carrierTracking.Events {
				// Check if event already exists
				eventExists := false
				for _, dbEvent := range events {
					if dbEvent.CarrierEventID == carrierEvent.CarrierEventID {
						eventExists = true
						break
					}
				}

				// Add new events
				if !eventExists {
					carrierEvent.ID = uuid.New().String()
					carrierEvent.ShipmentID = shipment.ID
					err := s.repo.CreateTrackingEvent(ctx, carrierEvent)
					if err != nil {
						s.logger.Error(ctx, "Failed to save tracking event", err)
					} else {
						events = append(events, carrierEvent)
					}
				}
			}

			// Update shipment status if changed
			if carrierTracking.Status != string(shipment.Status) {
				newStatus := models.ShipmentStatus(carrierTracking.Status)
				if newStatus.IsValid() && shipment.Status.CanTransitionTo(newStatus) {
					err := s.repo.UpdateShipmentStatus(ctx, shipment.ID, newStatus)
					if err != nil {
						s.logger.Error(ctx, "Failed to update shipment status", err)
					} else {
						shipment.Status = newStatus
					}
				}
			}
		}
	}

	// Convert events slice from []*ShipmentTrackingEvent to []ShipmentTrackingEvent
	eventsList := make([]models.ShipmentTrackingEvent, len(events))
	for i, event := range events {
		eventsList[i] = *event
	}

	trackingInfo := &models.TrackingInfo{
		TrackingNumber:    trackingNumber,
		CarrierTrackingID: shipment.CarrierTrackingID,
		Status:            shipment.Status,
		EstimatedDelivery: shipment.EstimatedDeliveryDate,
		ActualDelivery:    shipment.ActualDeliveryDate,
		Events:            eventsList,
		LastUpdated:       time.Now(),
	}

	return trackingInfo, nil
}

// UpdateShipmentStatus updates a shipment's status
func (s *ShippingService) UpdateShipmentStatus(ctx context.Context, id string, status models.ShipmentStatus) error {
	s.logger.Info(ctx, "Updating shipment status", map[string]interface{}{
		"shipment_id": id,
		"status":      status,
	})

	// Validate status
	if !status.IsValid() {
		return fmt.Errorf("invalid shipment status: %s", status)
	}

	// Get current shipment
	shipment, err := s.repo.GetShipment(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get shipment: %w", err)
	}

	// Check if transition is allowed
	if !shipment.Status.CanTransitionTo(status) {
		return fmt.Errorf("cannot transition from %s to %s", shipment.Status, status)
	}

	// Update status
	err = s.repo.UpdateShipmentStatus(ctx, id, status)
	if err != nil {
		s.logger.Error(ctx, "Failed to update shipment status", err, map[string]interface{}{
			"shipment_id": id,
			"status":      status,
		})
		return fmt.Errorf("failed to update shipment status: %w", err)
	}

	// Create tracking event
	event := &models.ShipmentTrackingEvent{
		ID:          uuid.New().String(),
		ShipmentID:  id,
		Status:      string(status),
		Description: fmt.Sprintf("Status updated to %s", status),
		EventTime:   time.Now(),
	}

	err = s.repo.CreateTrackingEvent(ctx, event)
	if err != nil {
		s.logger.Error(ctx, "Failed to create tracking event", err)
	}

	s.logger.Info(ctx, "Shipment status updated successfully", map[string]interface{}{
		"shipment_id": id,
		"new_status":  status,
	})

	return nil
}

// GetShippingMethods retrieves all active shipping methods
func (s *ShippingService) GetShippingMethods(ctx context.Context) ([]*models.ShippingMethod, error) {
	methods, err := s.repo.GetShippingMethods(ctx, true)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipping methods", err)
		return nil, fmt.Errorf("failed to get shipping methods: %w", err)
	}

	return methods, nil
}

// ValidateAddress validates a shipping address
func (s *ShippingService) ValidateAddress(ctx context.Context, address *models.Address) (*gateway.AddressValidationResponse, error) {
	s.logger.Info(ctx, "Validating address", map[string]interface{}{
		"country": address.Country,
		"state":   address.State,
	})

	// Try to find a carrier gateway for validation
	for _, carrierGateway := range s.gatewayFactory.GetAllGateways() {
		response, err := carrierGateway.ValidateAddress(ctx, address)
		if err == nil {
			return response, nil
		}
	}

	// Fallback to basic validation
	isValid := address.Street != "" && address.City != "" && 
			   address.State != "" && address.PostalCode != "" && address.Country != ""

	response := &gateway.AddressValidationResponse{
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

// Helper methods

// findApplicableRate finds the applicable shipping rate for weight and order value
func (s *ShippingService) findApplicableRate(rates []*models.ShippingRate, weightKg, orderValue decimal.Decimal) *models.ShippingRate {
	for _, rate := range rates {
		// Check weight constraints
		if rate.MaxWeightKg.IsPositive() && weightKg.GreaterThan(rate.MaxWeightKg) {
			continue
		}
		if weightKg.LessThan(rate.MinWeightKg) {
			continue
		}

		// Check order value constraints
		if rate.MaxOrderValue.IsPositive() && orderValue.GreaterThan(rate.MaxOrderValue) {
			continue
		}
		if orderValue.LessThan(rate.MinOrderValue) {
			continue
		}

		return rate
	}
	return nil
}

// calculateShippingCost calculates shipping cost based on method and request
func (s *ShippingService) calculateShippingCost(method *models.ShippingMethod, rate *models.ShippingRate, request *models.ShippingQuoteRequest) decimal.Decimal {
	cost := rate.Rate

	// Add weight-based cost
	if method.CostPerKg.IsPositive() {
		weightCost := request.WeightKg.Mul(method.CostPerKg)
		cost = cost.Add(weightCost)
	}

	return cost
}

// calculateShippingCostFromRate calculates shipping cost from rate and weight
func (s *ShippingService) calculateShippingCostFromRate(method *models.ShippingMethod, rate *models.ShippingRate, weightKg decimal.Decimal) decimal.Decimal {
	cost := rate.Rate

	// Add weight-based cost
	if method.CostPerKg.IsPositive() {
		weightCost := weightKg.Mul(method.CostPerKg)
		cost = cost.Add(weightCost)
	}

	return cost
}

// generateTrackingNumber generates a tracking number
func (s *ShippingService) generateTrackingNumber(carrier models.CarrierType) string {
	timestamp := time.Now().Unix()
	
	switch carrier {
	case models.CarrierFedEx:
		return fmt.Sprintf("FDX%d", timestamp)
	case models.CarrierUPS:
		return fmt.Sprintf("1Z%d", timestamp)
	case models.CarrierUSPS:
		return fmt.Sprintf("9400%d", timestamp)
	default:
		return fmt.Sprintf("SHP%d", timestamp)
	}
}

// createCarrierShipment creates shipment with carrier (async)
func (s *ShippingService) createCarrierShipment(ctx context.Context, shipment *models.Shipment) {
	// Get shipping method to determine carrier
	method, err := s.repo.GetShippingMethod(ctx, shipment.ShippingMethodID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get shipping method for carrier shipment", err)
		return
	}

	// Get carrier gateway
	gateway, exists := s.gatewayFactory.GetGateway(method.CarrierName)
	if !exists {
		s.logger.Info(ctx, "No carrier gateway available", map[string]interface{}{
			"carrier": method.CarrierName,
		})
		return
	}

	// Create shipment with carrier
	response, err := gateway.CreateShipment(ctx, shipment)
	if err != nil {
		s.logger.Error(ctx, "Failed to create carrier shipment", err)
		return
	}

	// Update shipment with carrier response
	shipment.CarrierTrackingID = response.CarrierTrackingID
	shipment.LabelURL = response.LabelURL
	shipment.EstimatedDeliveryDate = response.EstimatedDeliveryDate
	shipment.CarrierResponse = response.RawResponse
	shipment.Status = models.ShipmentProcessing

	err = s.repo.UpdateShipment(ctx, shipment)
	if err != nil {
		s.logger.Error(ctx, "Failed to update shipment with carrier response", err)
		return
	}

	s.logger.Info(ctx, "Carrier shipment created successfully", map[string]interface{}{
		"shipment_id":         shipment.ID,
		"carrier_tracking_id": response.CarrierTrackingID,
	})
}
