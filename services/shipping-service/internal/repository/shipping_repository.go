package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/shopsphere/shared/models"
)

// ShippingRepository defines the interface for shipping data operations
type ShippingRepository interface {
	// Shipping Methods
	CreateShippingMethod(ctx context.Context, method *models.ShippingMethod) error
	GetShippingMethod(ctx context.Context, id string) (*models.ShippingMethod, error)
	GetShippingMethods(ctx context.Context, active bool) ([]*models.ShippingMethod, error)
	UpdateShippingMethod(ctx context.Context, method *models.ShippingMethod) error

	// Shipping Zones
	GetShippingZoneByAddress(ctx context.Context, address *models.Address) (*models.ShippingZone, error)
	GetShippingZones(ctx context.Context, active bool) ([]*models.ShippingZone, error)

	// Shipping Rates
	GetShippingRates(ctx context.Context, methodID, zoneID string) ([]*models.ShippingRate, error)

	// Shipments
	CreateShipment(ctx context.Context, shipment *models.Shipment) error
	GetShipment(ctx context.Context, id string) (*models.Shipment, error)
	GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error)
	GetShipmentsByOrder(ctx context.Context, orderID string) ([]*models.Shipment, error)
	UpdateShipment(ctx context.Context, shipment *models.Shipment) error
	UpdateShipmentStatus(ctx context.Context, id string, status models.ShipmentStatus) error

	// Tracking Events
	CreateTrackingEvent(ctx context.Context, event *models.ShipmentTrackingEvent) error
	GetTrackingEvents(ctx context.Context, shipmentID string) ([]*models.ShipmentTrackingEvent, error)

	// Carrier Configs
	GetCarrierConfig(ctx context.Context, carrierName models.CarrierType) (*models.CarrierConfig, error)
}

// PostgresShippingRepository implements ShippingRepository using PostgreSQL
type PostgresShippingRepository struct {
	db *sql.DB
}

// NewPostgresShippingRepository creates a new PostgreSQL shipping repository
func NewPostgresShippingRepository(db *sql.DB) *PostgresShippingRepository {
	return &PostgresShippingRepository{db: db}
}

// CreateShippingMethod creates a new shipping method
func (r *PostgresShippingRepository) CreateShippingMethod(ctx context.Context, method *models.ShippingMethod) error {
	query := `
		INSERT INTO shipping_methods (id, name, description, carrier_name, service_type, 
			delivery_time_min, delivery_time_max, base_cost, cost_per_kg, cost_per_km, 
			max_weight_kg, max_dimensions_cm, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		method.ID, method.Name, method.Description, method.CarrierName, method.ServiceType,
		method.DeliveryTimeMin, method.DeliveryTimeMax, method.BaseCost, method.CostPerKg,
		method.CostPerKm, method.MaxWeightKg, method.MaxDimensions, method.IsActive,
	).Scan(&method.CreatedAt, &method.UpdatedAt)

	return err
}

// GetShippingMethod retrieves a shipping method by ID
func (r *PostgresShippingRepository) GetShippingMethod(ctx context.Context, id string) (*models.ShippingMethod, error) {
	method := &models.ShippingMethod{}
	query := `
		SELECT id, name, description, carrier_name, service_type, delivery_time_min, 
			delivery_time_max, base_cost, cost_per_kg, cost_per_km, max_weight_kg, 
			max_dimensions_cm, is_active, created_at, updated_at
		FROM shipping_methods WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&method.ID, &method.Name, &method.Description, &method.CarrierName, &method.ServiceType,
		&method.DeliveryTimeMin, &method.DeliveryTimeMax, &method.BaseCost, &method.CostPerKg,
		&method.CostPerKm, &method.MaxWeightKg, &method.MaxDimensions, &method.IsActive,
		&method.CreatedAt, &method.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return method, nil
}

// GetShippingMethods retrieves all shipping methods
func (r *PostgresShippingRepository) GetShippingMethods(ctx context.Context, active bool) ([]*models.ShippingMethod, error) {
	var query string

	if active {
		query = `
			SELECT id, name, description, carrier_name, service_type, delivery_time_min, 
				delivery_time_max, base_cost, cost_per_kg, cost_per_km, max_weight_kg, 
				max_dimensions_cm, is_active, created_at, updated_at
			FROM shipping_methods WHERE is_active = true ORDER BY name`
	} else {
		query = `
			SELECT id, name, description, carrier_name, service_type, delivery_time_min, 
				delivery_time_max, base_cost, cost_per_kg, cost_per_km, max_weight_kg, 
				max_dimensions_cm, is_active, created_at, updated_at
			FROM shipping_methods ORDER BY name`
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []*models.ShippingMethod
	for rows.Next() {
		method := &models.ShippingMethod{}
		err := rows.Scan(
			&method.ID, &method.Name, &method.Description, &method.CarrierName, &method.ServiceType,
			&method.DeliveryTimeMin, &method.DeliveryTimeMax, &method.BaseCost, &method.CostPerKg,
			&method.CostPerKm, &method.MaxWeightKg, &method.MaxDimensions, &method.IsActive,
			&method.CreatedAt, &method.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, rows.Err()
}

// UpdateShippingMethod updates a shipping method
func (r *PostgresShippingRepository) UpdateShippingMethod(ctx context.Context, method *models.ShippingMethod) error {
	query := `
		UPDATE shipping_methods 
		SET name = $2, description = $3, carrier_name = $4, service_type = $5,
			delivery_time_min = $6, delivery_time_max = $7, base_cost = $8, 
			cost_per_kg = $9, cost_per_km = $10, max_weight_kg = $11, 
			max_dimensions_cm = $12, is_active = $13, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		method.ID, method.Name, method.Description, method.CarrierName, method.ServiceType,
		method.DeliveryTimeMin, method.DeliveryTimeMax, method.BaseCost, method.CostPerKg,
		method.CostPerKm, method.MaxWeightKg, method.MaxDimensions, method.IsActive,
	).Scan(&method.UpdatedAt)

	return err
}

// GetShippingZoneByAddress finds the appropriate shipping zone for an address
func (r *PostgresShippingRepository) GetShippingZoneByAddress(ctx context.Context, address *models.Address) (*models.ShippingZone, error) {
	query := `
		SELECT id, name, description, countries, states, postal_codes, is_active, created_at, updated_at
		FROM shipping_zones 
		WHERE is_active = true 
		AND ($1 = ANY(countries) OR '*' = ANY(countries))
		ORDER BY 
			CASE WHEN $1 = ANY(countries) THEN 1 ELSE 2 END,
			CASE WHEN $2 = ANY(states) THEN 1 ELSE 2 END
		LIMIT 1`

	zone := &models.ShippingZone{}
	err := r.db.QueryRowContext(ctx, query, address.Country, address.State).Scan(
		&zone.ID, &zone.Name, &zone.Description, pq.Array(&zone.Countries),
		pq.Array(&zone.States), pq.Array(&zone.PostalCodes), &zone.IsActive,
		&zone.CreatedAt, &zone.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return zone, nil
}

// GetShippingZones retrieves all shipping zones
func (r *PostgresShippingRepository) GetShippingZones(ctx context.Context, active bool) ([]*models.ShippingZone, error) {
	var query string

	if active {
		query = `
			SELECT id, name, description, countries, states, postal_codes, is_active, created_at, updated_at
			FROM shipping_zones WHERE is_active = true ORDER BY name`
	} else {
		query = `
			SELECT id, name, description, countries, states, postal_codes, is_active, created_at, updated_at
			FROM shipping_zones ORDER BY name`
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*models.ShippingZone
	for rows.Next() {
		zone := &models.ShippingZone{}
		err := rows.Scan(
			&zone.ID, &zone.Name, &zone.Description, pq.Array(&zone.Countries),
			pq.Array(&zone.States), pq.Array(&zone.PostalCodes), &zone.IsActive,
			&zone.CreatedAt, &zone.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}

	return zones, rows.Err()
}

// GetShippingRates retrieves shipping rates by method and zone
func (r *PostgresShippingRepository) GetShippingRates(ctx context.Context, methodID, zoneID string) ([]*models.ShippingRate, error) {
	query := `
		SELECT id, shipping_method_id, shipping_zone_id, min_weight_kg, max_weight_kg, 
			min_order_value, max_order_value, rate, free_shipping_threshold, is_active, 
			created_at, updated_at
		FROM shipping_rates 
		WHERE is_active = true AND shipping_method_id = $1 AND shipping_zone_id = $2
		ORDER BY min_weight_kg, min_order_value`

	rows, err := r.db.QueryContext(ctx, query, methodID, zoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*models.ShippingRate
	for rows.Next() {
		rate := &models.ShippingRate{}
		err := rows.Scan(
			&rate.ID, &rate.ShippingMethodID, &rate.ShippingZoneID, &rate.MinWeightKg,
			&rate.MaxWeightKg, &rate.MinOrderValue, &rate.MaxOrderValue, &rate.Rate,
			&rate.FreeShippingThreshold, &rate.IsActive, &rate.CreatedAt, &rate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

// CreateShipment creates a new shipment
func (r *PostgresShippingRepository) CreateShipment(ctx context.Context, shipment *models.Shipment) error {
	fromAddrJSON, _ := json.Marshal(shipment.FromAddress)
	toAddrJSON, _ := json.Marshal(shipment.ToAddress)
	carrierResponseJSON, _ := json.Marshal(shipment.CarrierResponse)

	query := `
		INSERT INTO shipments (id, order_id, user_id, shipping_method_id, tracking_number, 
			carrier_tracking_id, status, from_address, to_address, weight_kg, dimensions_cm, 
			declared_value, shipping_cost, insurance_cost, total_cost, estimated_delivery_date, 
			actual_pickup_date, actual_delivery_date, carrier_response, label_url, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		shipment.ID, shipment.OrderID, shipment.UserID, shipment.ShippingMethodID,
		shipment.TrackingNumber, shipment.CarrierTrackingID, shipment.Status,
		fromAddrJSON, toAddrJSON, shipment.WeightKg, shipment.Dimensions,
		shipment.DeclaredValue, shipment.ShippingCost, shipment.InsuranceCost,
		shipment.TotalCost, shipment.EstimatedDeliveryDate, shipment.ActualPickupDate,
		shipment.ActualDeliveryDate, carrierResponseJSON, shipment.LabelURL, shipment.Notes,
	).Scan(&shipment.CreatedAt, &shipment.UpdatedAt)

	return err
}

// GetShipment retrieves a shipment by ID
func (r *PostgresShippingRepository) GetShipment(ctx context.Context, id string) (*models.Shipment, error) {
	shipment := &models.Shipment{}
	var fromAddrJSON, toAddrJSON, carrierResponseJSON []byte

	query := `
		SELECT id, order_id, user_id, shipping_method_id, tracking_number, carrier_tracking_id, 
			status, from_address, to_address, weight_kg, dimensions_cm, declared_value, 
			shipping_cost, insurance_cost, total_cost, estimated_delivery_date, 
			actual_pickup_date, actual_delivery_date, carrier_response, label_url, notes, 
			created_at, updated_at
		FROM shipments WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingMethodID,
		&shipment.TrackingNumber, &shipment.CarrierTrackingID, &shipment.Status,
		&fromAddrJSON, &toAddrJSON, &shipment.WeightKg, &shipment.Dimensions,
		&shipment.DeclaredValue, &shipment.ShippingCost, &shipment.InsuranceCost,
		&shipment.TotalCost, &shipment.EstimatedDeliveryDate, &shipment.ActualPickupDate,
		&shipment.ActualDeliveryDate, &carrierResponseJSON, &shipment.LabelURL,
		&shipment.Notes, &shipment.CreatedAt, &shipment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(fromAddrJSON, &shipment.FromAddress)
	json.Unmarshal(toAddrJSON, &shipment.ToAddress)
	json.Unmarshal(carrierResponseJSON, &shipment.CarrierResponse)

	return shipment, nil
}

// GetShipmentByTrackingNumber retrieves a shipment by tracking number
func (r *PostgresShippingRepository) GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	shipment := &models.Shipment{}
	var fromAddrJSON, toAddrJSON, carrierResponseJSON []byte

	query := `
		SELECT id, order_id, user_id, shipping_method_id, tracking_number, carrier_tracking_id, 
			status, from_address, to_address, weight_kg, dimensions_cm, declared_value, 
			shipping_cost, insurance_cost, total_cost, estimated_delivery_date, 
			actual_pickup_date, actual_delivery_date, carrier_response, label_url, notes, 
			created_at, updated_at
		FROM shipments WHERE tracking_number = $1`

	err := r.db.QueryRowContext(ctx, query, trackingNumber).Scan(
		&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingMethodID,
		&shipment.TrackingNumber, &shipment.CarrierTrackingID, &shipment.Status,
		&fromAddrJSON, &toAddrJSON, &shipment.WeightKg, &shipment.Dimensions,
		&shipment.DeclaredValue, &shipment.ShippingCost, &shipment.InsuranceCost,
		&shipment.TotalCost, &shipment.EstimatedDeliveryDate, &shipment.ActualPickupDate,
		&shipment.ActualDeliveryDate, &carrierResponseJSON, &shipment.LabelURL,
		&shipment.Notes, &shipment.CreatedAt, &shipment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(fromAddrJSON, &shipment.FromAddress)
	json.Unmarshal(toAddrJSON, &shipment.ToAddress)
	json.Unmarshal(carrierResponseJSON, &shipment.CarrierResponse)

	return shipment, nil
}

// GetShipmentsByOrder retrieves shipments by order ID
func (r *PostgresShippingRepository) GetShipmentsByOrder(ctx context.Context, orderID string) ([]*models.Shipment, error) {
	query := `
		SELECT id, order_id, user_id, shipping_method_id, tracking_number, carrier_tracking_id, 
			status, from_address, to_address, weight_kg, dimensions_cm, declared_value, 
			shipping_cost, insurance_cost, total_cost, estimated_delivery_date, 
			actual_pickup_date, actual_delivery_date, carrier_response, label_url, notes, 
			created_at, updated_at
		FROM shipments WHERE order_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []*models.Shipment
	for rows.Next() {
		shipment := &models.Shipment{}
		var fromAddrJSON, toAddrJSON, carrierResponseJSON []byte

		err := rows.Scan(
			&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingMethodID,
			&shipment.TrackingNumber, &shipment.CarrierTrackingID, &shipment.Status,
			&fromAddrJSON, &toAddrJSON, &shipment.WeightKg, &shipment.Dimensions,
			&shipment.DeclaredValue, &shipment.ShippingCost, &shipment.InsuranceCost,
			&shipment.TotalCost, &shipment.EstimatedDeliveryDate, &shipment.ActualPickupDate,
			&shipment.ActualDeliveryDate, &carrierResponseJSON, &shipment.LabelURL,
			&shipment.Notes, &shipment.CreatedAt, &shipment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		json.Unmarshal(fromAddrJSON, &shipment.FromAddress)
		json.Unmarshal(toAddrJSON, &shipment.ToAddress)
		json.Unmarshal(carrierResponseJSON, &shipment.CarrierResponse)

		shipments = append(shipments, shipment)
	}

	return shipments, rows.Err()
}

// UpdateShipment updates a shipment
func (r *PostgresShippingRepository) UpdateShipment(ctx context.Context, shipment *models.Shipment) error {
	fromAddrJSON, _ := json.Marshal(shipment.FromAddress)
	toAddrJSON, _ := json.Marshal(shipment.ToAddress)
	carrierResponseJSON, _ := json.Marshal(shipment.CarrierResponse)

	query := `
		UPDATE shipments 
		SET tracking_number = $2, carrier_tracking_id = $3, status = $4, 
			from_address = $5, to_address = $6, weight_kg = $7, dimensions_cm = $8, 
			declared_value = $9, shipping_cost = $10, insurance_cost = $11, 
			total_cost = $12, estimated_delivery_date = $13, actual_pickup_date = $14, 
			actual_delivery_date = $15, carrier_response = $16, label_url = $17, 
			notes = $18, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		shipment.ID, shipment.TrackingNumber, shipment.CarrierTrackingID, shipment.Status,
		fromAddrJSON, toAddrJSON, shipment.WeightKg, shipment.Dimensions,
		shipment.DeclaredValue, shipment.ShippingCost, shipment.InsuranceCost,
		shipment.TotalCost, shipment.EstimatedDeliveryDate, shipment.ActualPickupDate,
		shipment.ActualDeliveryDate, carrierResponseJSON, shipment.LabelURL, shipment.Notes,
	).Scan(&shipment.UpdatedAt)

	return err
}

// UpdateShipmentStatus updates a shipment's status
func (r *PostgresShippingRepository) UpdateShipmentStatus(ctx context.Context, id string, status models.ShipmentStatus) error {
	query := `UPDATE shipments SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// CreateTrackingEvent creates a new tracking event
func (r *PostgresShippingRepository) CreateTrackingEvent(ctx context.Context, event *models.ShipmentTrackingEvent) error {
	carrierRawDataJSON, _ := json.Marshal(event.CarrierRawData)

	query := `
		INSERT INTO shipment_tracking_events (id, shipment_id, status, description, location, 
			event_time, carrier_event_id, carrier_raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.db.QueryRowContext(ctx, query,
		event.ID, event.ShipmentID, event.Status, event.Description, event.Location,
		event.EventTime, event.CarrierEventID, carrierRawDataJSON,
	).Scan(&event.CreatedAt)

	return err
}

// GetTrackingEvents retrieves tracking events for a shipment
func (r *PostgresShippingRepository) GetTrackingEvents(ctx context.Context, shipmentID string) ([]*models.ShipmentTrackingEvent, error) {
	query := `
		SELECT id, shipment_id, status, description, location, event_time, 
			carrier_event_id, carrier_raw_data, created_at
		FROM shipment_tracking_events 
		WHERE shipment_id = $1 ORDER BY event_time DESC`

	rows, err := r.db.QueryContext(ctx, query, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.ShipmentTrackingEvent
	for rows.Next() {
		event := &models.ShipmentTrackingEvent{}
		var carrierRawDataJSON []byte

		err := rows.Scan(
			&event.ID, &event.ShipmentID, &event.Status, &event.Description,
			&event.Location, &event.EventTime, &event.CarrierEventID,
			&carrierRawDataJSON, &event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON field
		json.Unmarshal(carrierRawDataJSON, &event.CarrierRawData)

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetCarrierConfig retrieves carrier configuration
func (r *PostgresShippingRepository) GetCarrierConfig(ctx context.Context, carrierName models.CarrierType) (*models.CarrierConfig, error) {
	config := &models.CarrierConfig{}
	var configDataJSON []byte

	query := `
		SELECT id, carrier_name, api_endpoint, api_key_encrypted, api_secret_encrypted, 
			webhook_secret, is_active, rate_limit_per_minute, timeout_seconds, 
			config_data, created_at, updated_at
		FROM carrier_configs WHERE carrier_name = $1 AND is_active = true`

	err := r.db.QueryRowContext(ctx, query, carrierName).Scan(
		&config.ID, &config.CarrierName, &config.APIEndpoint, &config.APIKeyEncrypted,
		&config.APISecretEncrypted, &config.WebhookSecret, &config.IsActive,
		&config.RateLimitPerMinute, &config.TimeoutSeconds, &configDataJSON,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON field
	json.Unmarshal(configDataJSON, &config.ConfigData)

	return config, nil
}
