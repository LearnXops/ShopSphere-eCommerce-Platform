module github.com/shopsphere/notification-service

go 1.21

require (
	github.com/google/uuid v1.5.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/shopsphere/shared v0.0.0
)

require github.com/sendgrid/rest v2.6.9+incompatible // indirect

replace github.com/shopsphere/shared => ../../shared
