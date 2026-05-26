package protocol

type PayType string

const (
	PayTypeNative PayType = "native"
	PayTypeJSAPI  PayType = "jsapi"
	PayTypeH5     PayType = "h5"
)

var validPayTypes = map[PayType]bool{
	PayTypeNative: true,
	PayTypeJSAPI:  true,
	PayTypeH5:     true,
}

func (t PayType) IsValid() bool {
	return validPayTypes[t]
}

func (t PayType) String() string {
	return string(t)
}
