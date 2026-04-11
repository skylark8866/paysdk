module shop-demo

go 1.21

require xgdn-pay v0.0.0

require (
	github.com/go-resty/resty/v2 v2.12.0 // indirect
	golang.org/x/net v0.24.0 // indirect
)

replace xgdn-pay => ../..
