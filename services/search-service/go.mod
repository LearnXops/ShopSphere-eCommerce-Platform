module github.com/shopsphere/search-service

go 1.21

require (
	github.com/shopsphere/shared v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/elastic/go-elasticsearch/v8 v8.11.1
)

replace github.com/shopsphere/shared => ../../shared