module github.com/horizoonn/factory-platform.git/payment

go 1.26.2

require (
	github.com/google/uuid v1.6.0
	github.com/horizoonn/factory-platform.git/shared v0.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	google.golang.org/grpc v1.81.0
)

require (
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/horizoonn/factory-platform.git/shared => ../shared
