module github.com/shopsphere/shipping-service

go 1.21

require (
	github.com/shopsphere/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/google/uuid v1.6.0
	github.com/shopspring/decimal v1.3.1
)

replace github.com/shopsphere/shared => ../../shared