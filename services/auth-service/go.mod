module github.com/shopsphere/auth-service

go 1.21

require (
	github.com/gorilla/mux v1.8.1
	github.com/shopsphere/shared v0.0.0
)

require github.com/google/uuid v1.5.0 // indirect

replace github.com/shopsphere/shared => ../../shared
