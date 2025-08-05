module github.com/shopsphere/notification-service

go 1.21

require (
	github.com/shopsphere/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
)

replace github.com/shopsphere/shared => ../../shared