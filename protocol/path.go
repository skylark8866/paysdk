package protocol

const (
	PathOrderCreate = "/api/v1/order/create"
	PathOrderQuery  = "/api/v1/order/query"
	PathOrderCheck  = "/api/v1/order/check"
	PathOrderClose  = "/api/v1/order/close"
	PathRefund      = "/api/v1/refund"
	PathRefundQuery = "/api/v1/refund/query"
	PathRefundOrder = "/api/v1/refund/order"
	PathRefundInfo  = "/api/v1/refund/info"
	PathSSESubscribe = "/api/v1/sse/subscribe/{orderNo}"
)
