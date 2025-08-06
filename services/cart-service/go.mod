module github.com/shopsphere/cart-service

go 1.21

require (
	github.com/gorilla/mux v1.8.1
	github.com/redis/go-redis/v9 v9.3.0
	github.com/shopsphere/shared v0.0.0
	github.com/shopspring/decimal v1.3.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
)

replace github.com/shopsphere/shared => ../../shared
