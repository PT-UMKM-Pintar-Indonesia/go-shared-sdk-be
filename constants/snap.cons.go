package sdk_cons

const (
	TRANSACTION_LATEST_CODE_SUCCESS   = "00"
	TRANSACTION_LATEST_CODE_INITIATED = "01"
	TRANSACTION_LATEST_CODE_PAYING    = "02"
	TRANSACTION_LATEST_CODE_PENDING   = "03"
	TRANSACTION_LATEST_CODE_REFUNDED  = "04"
	TRANSACTION_LATEST_CODE_CANCELED  = "05"
	TRANSACTION_LATEST_CODE_FAILED    = "06"
	TRANSACTION_LATEST_CODE_EXPIRED   = "07" // Ditambahkan: Transaksi kedaluwarsa
)

// --- VIRTUAL ACCOUNT ---
const (
	VA_STATUS_SUCCESS_CODE  = "00" // Success: Transaksi sudah berhasil terbayar
	VA_STATUS_PENDING_CODE  = "03" // Pending: Transaksi yang belum terbayar
	VA_STATUS_REFUNDED_CODE = "04" // Refunded: Transaksi yang sudah di-refund
	VA_STATUS_CANCELED_CODE = "05" // Canceled: Transaksi dibatalkan secara aktif oleh merchant/sistem
	VA_STATUS_FAILED_CODE   = "06" // Failed: Transaksi yang gagal
	VA_STATUS_EXPIRED_CODE  = "07" // Ditambahkan - Expired: Waktu pembayaran VA telah habis

	VA_STATUS_SUCCESS  = "success"
	VA_STATUS_PENDING  = "pending"
	VA_STATUS_REFUNDED = "refunded"
	VA_STATUS_CANCELED = "canceled"
	VA_STATUS_FAILED   = "failed"
	VA_STATUS_EXPIRED  = "expired" // Ditambahkan
)

// --- E-WALLET ---
const (
	EWALLET_STATUS_SUCCESS_CODE    = "00" // Success: Transaksi sudah berhasil terbayar
	EWALLET_STATUS_INITIATED_CODE  = "01" // Initiated: Transaksi yang sedang dalam proses inisiasi untuk dibayarkan
	EWALLET_STATUS_PENDING_CODE    = "03" // Pending: Transaksi yang belum terbayar
	EWALLET_STATUS_REFUNDED_CODE   = "04" // Refund: Transaksi yang sudah di-refund
	EWALLET_STATUS_CANCELED_CODE   = "05" // Canceled: Transaksi dibatalkan
	EWALLET_STATUS_FAILED_CODE     = "06" // Failed: Transaksi yang gagal
	EWALLET_STATUS_EXPIRED_CODE    = "07" // Ditambahkan - Expired: Timeout/Kedaluwarsa karena user tidak konfirmasi di app E-Wallet
	EWALLET_STATUS_CHARGEBACK_CODE = "08" // Ditambahkan (Opsional) - Chargeback: User komplain & dana ditarik e-wallet

	EWALLET_STATUS_SUCCESS    = "success"
	EWALLET_STATUS_INITIATED  = "initiated"
	EWALLET_STATUS_PENDING    = "pending"
	EWALLET_STATUS_REFUNDED   = "refund"
	EWALLET_STATUS_CANCELED   = "canceled"
	EWALLET_STATUS_FAILED     = "failed"
	EWALLET_STATUS_EXPIRED    = "expired"    // Ditambahkan
	EWALLET_STATUS_CHARGEBACK = "chargeback" // Ditambahkan
)

// --- QRIS ---
const (
	QRIS_STATUS_SUCCESS_CODE  = "00" // Success: Transaksi sudah berhasil terbayar
	QRIS_STATUS_PENDING_CODE  = "03" // Pending: Transaksi yang belum terbayar
	QRIS_STATUS_REFUNDED_CODE = "04" // Refund: Transaksi yang sudah di-refund
	QRIS_STATUS_CANCELED_CODE = "05" // Canceled: Transaksi dibatalkan
	QRIS_STATUS_FAILED_CODE   = "06" // Failed: Transaksi yang gagal
	QRIS_STATUS_EXPIRED_CODE  = "07" // Ditambahkan - Expired: QRIS dinamis melewati batas waktu bayar

	QRIS_STATUS_SUCCESS  = "success"
	QRIS_STATUS_PENDING  = "pending"
	QRIS_STATUS_REFUNDED = "refund"
	QRIS_STATUS_CANCELED = "canceled"
	QRIS_STATUS_FAILED   = "failed"
	QRIS_STATUS_EXPIRED  = "expired" // Ditambahkan
)

// --- DISBURSEMENT (PENGIRIMAN DANA) ---
const (
	DISBURSEMENT_STATUS_SUCCESS_CODE    = "00" // Success: Transaksi berhasil dan dana masuk ke rekening tujuan
	DISBURSEMENT_STATUS_PENDING_CODE    = "03" // Pending: Menunggu proses approval atau masuk antrean
	DISBURSEMENT_STATUS_CANCELED_CODE   = "04" // Canceled: Transaksi disbursement terjadwal yang dibatalkan
	DISBURSEMENT_STATUS_REJECTED_CODE   = "05" // Rejected: Transaksi yang di-reject oleh sistem/approver
	DISBURSEMENT_STATUS_SUSPECT_CODE    = "08" // Suspect: Transaksi status belum pasti (timeout ke bank)
	DISBURSEMENT_STATUS_FAILED_CODE     = "09" // Failed: Transaksi disbursement yang gagal
	DISBURSEMENT_STATUS_PROCESSING_CODE = "10" // Ditambahkan - Processing: Sedang diproses oleh bank/jaringan BI-FAST
	DISBURSEMENT_STATUS_RETURNED_CODE   = "11" // Ditambahkan - Returned: Dana retur dari bank (rekening tujuan pasif/tutup)

	DISBURSEMENT_STATUS_SUCCESS    = "success"
	DISBURSEMENT_STATUS_PENDING    = "pending"
	DISBURSEMENT_STATUS_CANCELED   = "canceled"
	DISBURSEMENT_STATUS_REJECTED   = "rejected"
	DISBURSEMENT_STATUS_SUSPECT    = "suspect"
	DISBURSEMENT_STATUS_FAILED     = "failed"
	DISBURSEMENT_STATUS_PROCESSING = "processing" // Ditambahkan
	DISBURSEMENT_STATUS_RETURNED   = "returned"   // Ditambahkan
)

const (
	// Success (200)
	SNAP_SUCCESS = "20000"

	// Bad Request (400)
	SNAP_BAD_REQUEST             = "40000" // Added: Generic Bad Request
	SNAP_INVALID_FIELD_FORMAT    = "40001"
	SNAP_INVALID_MANDATORY_FIELD = "40002"
	SNAP_INVALID_DATE_TIME       = "40003" // Added: Invalid Date/Time format
	SNAP_INVALID_AMOUNT          = "40004" // Added: Amount format invalid or mismatched

	// Unauthorized (401)
	SNAP_UNAUTHORIZED         = "40100"
	SNAP_ACCESS_TOKEN_INVALID = "40101"
	SNAP_INVALID_CLIENT_ID    = "40111" // Added: Invalid App/Client/Partner ID
	SNAP_INVALID_SIGNATURE    = "40173" // Added: Invalid Signature (Symmetric/Asymmetric) - Sangat wajib di SNAP BI

	// Forbidden (403)
	SNAP_FORBIDDEN                = "40300" // Added: Generic Forbidden
	SNAP_FEATURE_NOT_ALLOWED      = "40301"
	SNAP_EXCEEDS_TXN_AMOUNT_LIMIT = "40302"
	SNAP_SUSPECTED_FRAUD          = "40303" // Added: Suspected fraud / Risk rejection
	SNAP_DO_NOT_HONOR             = "40305"
	SNAP_TXN_CANCELLED            = "40306" // Added: Transaction cancelled by user/system
	SNAP_INSUFFICIENT_FUNDS       = "40314"
	SNAP_TXN_NOT_PERMITTED        = "40315"
	SNAP_ACCOUNT_BLOCKED          = "40316" // Added: Account is blocked/suspended
	SNAP_TXN_EXPIRED              = "40317" // Added: Transaction expired
	SNAP_ACCOUNT_INACTIVE         = "40318"
	SNAP_SET_LIMIT_NOT_ALLOWED    = "40321"
	SNAP_ACCOUNT_LIMIT_EXCEED     = "40323"

	// Not Found (404)
	SNAP_NOT_FOUND            = "40400" // Added: Generic Not Found
	SNAP_TXN_NOT_FOUND        = "40401" // Added: Transaction reference not found
	SNAP_INVALID_ACCOUNT      = "40411"
	SNAP_INVALID_CUSTOMER     = "40412" // Added: Customer/Biller not found
	SNAP_PAID_BILL            = "40413"
	SNAP_BILL_EXPIRED         = "40414" // Added: Bill/VA already expired
	SNAP_INVALID_VIRTUAL_ACCT = "40419"

	// Conflict (409)
	SNAP_CONFLICT                 = "40900"
	SNAP_DUPLICATE_PARTNER_REF_NO = "40901"
	SNAP_DUPLICATE_REQUEST        = "40902" // Added: Idempotency key conflict

	// Server Errors (500/503/504)
	SNAP_GENERAL_ERROR       = "50000"
	SNAP_SERVICE_UNAVAILABLE = "50300" // Added: Maintenance / Down system
	SNAP_REQUEST_TIMEOUT     = "50400"
)

var SnapResponseMessageError = map[string]string{
	// 200 Series
	// "20000": "Transaction successful",

	// 400 Series
	"40000": "Network: Bad request; please check your request payload",
	"40001": "Network: The submitted data format is incorrect",
	"40002": "Network: Required information is missing",
	"40003": "Network: Invalid date or time format",
	"40004": "Network: Transaction amount is invalid or mismatched",

	// 401 Series
	"40100": "Network: Authorization failed; please check your credentials",
	"40101": "Network: Access token is invalid or has expired",
	"40111": "Network: Client ID or Partner ID is invalid",
	"40173": "Network: Invalid signature; digital signature mismatch",

	// 403 Series
	"40300": "Network: Access to the requested resource is forbidden",
	"40301": "Network: This feature is not available for your account",
	"40302": "Network: Transaction amount exceeds the permitted limit",
	"40303": "Network: Transaction declined due to suspected fraud",
	"40305": "Network: Transaction declined by the bank (Do Not Honor)",
	"40306": "Network: Transaction has been cancelled",
	"40314": "Network: Insufficient funds in your account",
	"40315": "Network: Transaction not permitted for this account",
	"40316": "Network: Account is currently blocked or suspended",
	"40317": "Network: Transaction has expired",
	"40318": "Network: Account is currently inactive",
	"40321": "Network: Limit adjustment cannot be processed at this time",
	"40323": "Network: Daily transaction limit has been reached",

	// 404 Series
	"40400": "Network: The requested resource was not found",
	"40401": "Network: Transaction reference not found",
	"40411": "Network: Account number is invalid or not found",
	"40412": "Network: Customer or biller information not found",
	"40413": "Network: This bill has already been paid",
	"40414": "Network: This bill or virtual account has expired",
	"40419": "Network: Virtual Account number is invalid",

	// 409 Series
	"40900": "Network: A data conflict occurred with the request",
	"40901": "Network: Transaction with this reference has already been processed",
	"40902": "Network: Duplicate request detected (Idempotency conflict)",

	// 500+ Series
	"50000": "Network: A system error occurred; please try again later",
	"50300": "Network: Service is currently unavailable or under maintenance",
	"50400": "Network: Request timed out; please try again",
}
