module github.com/shopsphere/auth-service

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/shopsphere/shared v0.0.0
	golang.org/x/crypto v0.17.0
)

require (
	github.com/google/uuid v1.5.0
	github.com/shopspring/decimal v1.3.1 // indirect
)

replace github.com/shopsphere/shared => ../../shared
