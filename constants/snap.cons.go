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
	// "20000": "Successful",
	// "20200": "Request In Progress",
	// "2006200" "2006200",

	"40000": "Bad Request",
	"40001": "Invalid Field Format {field name}",
	"40002": "Invalid Mandatory Field {field name}",
	"40100": "Unauthorized [reason]",
	"40101": "Invalid Token (B2B)",
	"40102": "Invalid Customer Token",
	"40103": "Token Not Found (B2B)",
	"40104": "Customer Token Not Found",
	"40300": "Transaction Expired",
	"40301": "Feature Not Allowed [Reason]",
	"40302": "Exceeds Transaction Amount Limit",
	"40303": "Suspected Fraud",
	"40304": "Activity Count Limit Exceeded",
	"40305": "Do Not Honor",
	"40306": "Feature Not Allowed At This Time [reason]",
	"40307": "Card Blocked",
	"40308": "Card Expired",
	"40309": "Dormant Account",
	"40310": "Need To Set Token Limit",
	"40311": "OTP Blocked",
	"40312": "OTP Lifetime Expired",
	"40313": "OTP Sent To Cardholder",
	"40314": "Insufficient Funds",
	"40315": "Transaction Not Permitted [reason]",
	"40316": "Suspend Transaction",
	"40317": "Token Limit Exceeded",
	"40318": "Inactive Card/Account/Customer",
	"40319": "Merchant Blacklisted",
	"40320": "Merchant Limit Exceed",
	"40321": "Set Limit Not Allowed",
	"40322": "Token Limit Invalid",
	"40323": "Account Limit Exceed",
	"40324": "Invalid CVV",
	"40325": "Invalid Expiration",
	"40326": "Invalid Card Type",
	"40327": "Card Usage Exceeded",
	"40328": "Restricted by CVM",
	"40329": "Duplicate Authorization",
	"40330": "Error OTP Request",
	"40331": "EMI Not Available For This Card/Bank Issuer",
	"40332": "Payer Is Invalid",
	"40333": "Account No Is Invalid",
	"40400": "Invalid Transaction Status",
	"40401": "Transaction Not Found",
	"40402": "Invalid Routing",
	"40403": "Bank Not Supported By Switch",
	"40404": "Transaction Cancelled",
	"40405": "Merchant Is Not Registered For Card Registration Services",
	"40406": "Need To Request OTP",
	"40407": "Journey Not Found",
	"40408": "Invalid Merchant",
	"40409": "No Issuer",
	"40410": "Invalid API Transition",
	"40411": "Invalid Card/Account/Customer [info]/Virtual Account",
	"40412": "Invalid Bill/Virtual Account [Reason]",
	"40413": "Invalid Amount",
	"40414": "Paid Bill",
	"40415": "Invalid OTP",
	"40416": "Partner Not Found",
	"40417": "Invalid Terminal",
	"40418": "Inconsistent Request",
	"40419": "Invalid Bill/Virtual Account (Expired)",
	"40500": "Requested Function Is Not Supported",
	"40501": "Requested Operation Is Not Allowed",
	"40900": "Conflict",
	"40901": "Duplicate partnerReferenceNo",
	"40902": "Transaction Has Been Processed",
	"42900": "Too Many Requests",
	"50000": "General Error",
	"50001": "Internal Server Error",
	"50002": "External Server Error",
	"50400": "Timeout",
}
