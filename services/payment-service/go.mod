module github.com/shopsphere/payment-service

go 1.21

require (
	github.com/shopsphere/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/stripe/stripe-go/v76 v76.16.0
)

replace github.com/shopsphere/shared => ../../shared