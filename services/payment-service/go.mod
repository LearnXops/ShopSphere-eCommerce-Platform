module github.com/shopsphere/payment-service

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/shopsphere/shared v0.0.0
	github.com/shopspring/decimal v1.3.1
	github.com/stripe/stripe-go/v76 v76.25.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
)

replace github.com/shopsphere/shared => ../../shared
