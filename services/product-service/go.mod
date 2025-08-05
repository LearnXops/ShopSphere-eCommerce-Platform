module github.com/shopsphere/product-service

go 1.21

require (
	github.com/elastic/go-elasticsearch/v8 v8.11.1
	github.com/google/uuid v1.5.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/shopsphere/shared v0.0.0
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/shopsphere/shared => ../../shared
