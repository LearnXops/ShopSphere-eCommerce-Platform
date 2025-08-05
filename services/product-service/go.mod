module github.com/shopsphere/product-service

go 1.21

require (
	github.com/shopsphere/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	go.mongodb.org/mongo-driver v1.13.1
)

replace github.com/shopsphere/shared => ../../shared