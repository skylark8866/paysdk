module xgdn-pay-example

go 1.21

require xgdn-pay v0.0.0

replace xgdn-pay => ../

require github.com/go-resty/resty/v2 v2.12.0 // indirect

require golang.org/x/net v0.24.0 // indirect
