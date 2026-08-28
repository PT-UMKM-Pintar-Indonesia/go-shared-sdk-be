package sdk_cons

import "runtime"

const (
	X_RABBIT_SECRET        = "x-rabbit-secret"
	X_RABBIT_UNKNOWN       = "x-rabbit-unknown"
	X_RABBIT_QUEUE         = "x-rabbit-queue"
	X_RABBIT_EXCHANGE      = "x-rabbit-exchange"
	X_RABBIT_EXCHANGE_TYPE = "x-rabbit-exchange-type"
	X_MESSAGE_TTL          = "x-message-ttl"
)

const (
	EXCHANGE_TYPE_DIRECT  = "direct"
	EXCHANGE_TYPE_FANOUT  = "fanout"
	EXCHANGE_TYPE_TOPIC   = "topic"
	EXCHANGE_TYPE_HEADERS = "headers"
)

const (
	EXCHANGE_NAME_DIRECT  = "amq.direct"
	EXCHANGE_NAME_TOPIC   = "amq.topic"
	EXCHANGE_NAME_FANOUT  = "amq.fanout"
	EXCHANGE_NAME_HEADERS = "amq.headers"
)

const (
	RPC  = "rpc"
	NRPC = "nrpc"
)

var (
	RABBITMQ_CONCURRENCY          int   = int(float64(runtime.NumCPU()) * 0.75)
	RABBITMQ_PREFETCH_FAST        int   = 20
	RABBITMQ_PREFETCH_SLOW        int   = 1
	RABBITMQ_SEMAPHORE_FAST       int64 = 1
	RABBITMQ_SEMAPHORE_SLOW       int64 = 5
	RABBITMQ_INVALID_REQUEST_BODY       = "Invalid request body"
	RABBITMQ_SERVICE_BUSY               = "Service is busy, please try again later"
)

const (
	RABBITMQ_EXCHANGE_NAME_EXTERNAL = "ex.worker.external"
	RABBITMQ_EXCHANGE_NAME_INTERNAL = "ex.worker.internal"
	RABBITMQ_EXCHANGE_NAME_PARTNER  = "ex.worker.partner"
)

const (
	RABBITMQ_QUEUE_WORKER_BALANCE                       = "worker.balances"
	RABBITMQ_QUEUE_WORKER_CALLBACK_DEPOSITS             = "worker.callback.deposits"
	RABBITMQ_QUEUE_WORKER_CALLBACK_DISBURSEMENTS        = "worker.callback.disbursements"
	RABBITMQ_QUEUE_WORKER_PARTNER_DEPOSITS              = "worker.partner.deposits"
	RABBITMQ_QUEUE_WORKER_PARTNER_CALLBACK_DEPOSITS     = "worker.partner.callback.deposits"
	RABBITMQ_QUEUE_WORKER_PARTNER_CALLBACK_DISBURSMENTS = "worker.partner.callback.disbursements"
	RABBITMQ_QUEUE_WORKER_EXPORT_INTERNAL               = "worker.export.internal"
	RABBITMQ_QUEUE_WORKER_EXPORT_EXTERNAL               = "worker.export.external"

	RABBITMQ_QUEUE_WORKER_FCM                              = "worker.fcm"
	RABBITMQ_QUEUE_WORKER_MAILER                           = "worker.mailer"
	RABBITMQ_QUEUE_WORKER_PRODUCT_INJECTION                = "worker.product.injection"
	RABBITMQ_QUEUE_WORKER_CALLBACK_TRANSFER_MOBILE         = "worker.callback.transfer.mobile"
	RABBITMQ_QUEUE_WORKER_CALLBACK_MOBILE_PAYMENT_DEPOSITS = "worker.callback.mobile.payment.deposits"
	RABBITMQ_QUEUE_WORKER_CALLBACK_MOBILE_STORE_DEPOSITS   = "worker.callback.mobile.store.deposits"
	RABBITMQ_QUEUE_WORKER_MOBILE_PRODUCT_INJECTION         = "worker.mobile.product.injection"
	RABBITMQ_QUEUE_WORKER_MOBILE_PRODUCT_PROCESSING        = "worker.mobile.product.processing"
	RABBITMQ_QUEUE_WORKER_SETTLEMENT                       = "worker.settlement"
	RABBITMQ_QUEUE_WORKER_RESETTLEMENT                     = "worker.resettlement"
	RABBITMQ_QUEUE_WORKER_CALLBACK_SINGLE_RETRIES          = "worker.callback.single.retries"
	RABBITMQ_QUEUE_WORKER_CALLBACK_BULK_RETRIES            = "worker.callback.bulk.retries"
	RABBITMQ_QUEUE_WORKER_PUT_OBJECT                       = "worker.upload.putobject"
	RABBITMQ_QUEUE_WORKER_PRESIGNED_URL                    = "worker.upload.presignedurl"
	RABBITMQ_QUEUE_WORKER_TELEGRAM_REPORT_TRANSACTION      = "worker.telegram.report.transaction"
	RABBITMQ_QUEUE_WORKER_TELEGRAM_BALANCE_MONITORING      = "worker.telegram.balance.monitoring"

	RABBITMQ_QUEUE_WORKER_NETZME_QRIS           = "worker.netzme.qris"
	RABBITMQ_QUEUE_WORKER_NETZME_QRIS_STATIC    = "worker.netzme.qris.static"
	RABBITMQ_QUEUE_WORKER_NETZME_INQUIRY_STATUS = "worker.netzme.inquiry.status"

	RABBITMQ_QUEUE_WORKER_PAKAILINK_INQUIRY      = "worker.pakailink.inquiry"
	RABBITMQ_QUEUE_WORKER_PAKAILINK_DISBURSEMENT = "worker.pakailink.disbursement"
	RABBITMQ_QUEUE_WORKER_PAKAILINK_QRIS         = "worker.pakailink.qris"
	RABBITMQ_QUEUE_WORKER_PAKAILINK_VA           = "worker.pakailink.va"

	RABBITMQ_QUEUE_WORKER_SIMULATOR_INQUIRY      = "worker.simulator.inquiry"
	RABBITMQ_QUEUE_WORKER_SIMULATOR_DISBURSEMENT = "worker.simulator.disbursement"
	RABBITMQ_QUEUE_WORKER_SIMULATOR_QRIS         = "worker.simulator.qris"
	RABBITMQ_QUEUE_WORKER_SIMULATOR_VA           = "worker.simulator.va"

	RABBITMQ_QUEUE_WORKER_SAEBO_INQUIRY        = "worker.saebo.inquiry"
	RABBITMQ_QUEUE_WORKER_SAEBO_TRANSFER       = "worker.saebo.disbursement"
	RABBITMQ_QUEUE_WORKER_SAEBO_QRIS           = "worker.saebo.qris"
	RABBITMQ_QUEUE_WORKER_SAEBO_VA             = "worker.saebo.va"
	RABBITMQ_QUEUE_WORKER_SAEBO_INQUIRY_STATUS = "worker.saebo.inquiry.status"

	RABBITMQ_QUEUE_WORKER_DOKU_INQUIRY  = "worker.doku.inquiry"
	RABBITMQ_QUEUE_WORKER_DOKU_TRANSFER = "worker.doku.disbursement"
	RABBITMQ_QUEUE_WORKER_DOKU_QRIS     = "worker.doku.qris"
	RABBITMQ_QUEUE_WORKER_DOKU_VA       = "worker.doku.va"

	RABBITMQ_QUEUE_WORKER_SINGAPAY_INQUIRY      = "worker.singapay.inquiry"
	RABBITMQ_QUEUE_WORKER_SINGAPAY_DISBURSEMENT = "worker.singapay.disbursement"
	RABBITMQ_QUEUE_WORKER_SINGAPAY_QRIS         = "worker.singapay.qris"
	RABBITMQ_QUEUE_WORKER_SINGAPAY_VA           = "worker.singapay.va"

	RABBITMQ_QUEUE_WORKER_YUK_INQUIRY      = "worker.yuk.inquiry"
	RABBITMQ_QUEUE_WORKER_YUK_DISBURSEMENT = "worker.yuk.disbursement"
	RABBITMQ_QUEUE_WORKER_YUK_QRIS         = "worker.yuk.qris"
	RABBITMQ_QUEUE_WORKER_YUK_VA           = "worker.yuk.va"
)
