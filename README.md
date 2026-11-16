# Shared SDK

Repository ini berisi shared SDK yang akan digunakan untuk pegembangan aplikasi dan bersifat reusable tools.

# Install Dependency

- ###  Install Latest Version
```sh
go get github.com/PT-UMKM-Pintar-Indonesia/shared-sdk
```

- ### Install Specific Version
```sh
go get github.com/PT-UMKM-Pintar-Indonesia/shared-sdk@v0.0.1
```

## Logging

`sdk_pkg.InitLogger` mengambil alih formatter global `logrus` dan menempelkan
label yang dipromosikan Cloud Logging menjadi `LogEntry.labels` dan
`LogEntry.severity`. Panggil sebagai statement pertama di `main`, sebelum ada
call site log lain:

```go
func main() {
	sdk_pkg.InitLogger(&sdk_dto.LoggerOptions{})
	// ...
}
```

Semua field `LoggerOptions` opsional; yang kosong jatuh ke environment
variable, lalu ke nilai turunan:

| Label | Option | Environment | Default |
|---|---|---|---|
| `service` | `Service` | `LOG_SERVICE_NAME` | nama binary |
| `host` | `Host` | `LOG_SERVICE_HOST` | hostname (di GCE = nama instance) |
| `instance` | `Instance` | `LOG_SERVICE_INSTANCE` | — |
| `zone` | `Zone` | `LOG_SERVICE_ZONE` | — |
| `env` | `Environment` | `APP_ENV`, `ENV`, `GIN_MODE` | `unknown` |
| — | `Level` | `LOG_LEVEL` | `info` |
| — | `Format` | `LOG_FORMAT` | `json` (`text` untuk lokal) |

Satu service yang jalan di beberapa VM tetap satu `labels.service`, sementara
`labels.host` dan `labels.instance` membedakan replika.

`sdk_pkg.Logrus` tetap bisa dipakai seperti sebelumnya. Formatter lazy-nya
sekarang mundur begitu `InitLogger` berjalan, jadi caller tidak perlu akal-akalan
untuk mempertahankan formatter sendiri. Caller yang belum memakai `InitLogger`
tetap mendapat perilaku lama: text, kecuali `LOG_FORMAT=json`.

## Code Of Conduct

Jika ingin menambahkan fungsi baru pastikan sesuai dengan tempat nya masing - masing, apakah itu **helpers**, **pkg**, **connection** dan sebagainya. Berikut adalah contoh saya menambahkan **helpers** `ToByte` berarti nanti fungsi tersebut akan masuk ke **helpers -> parser.helper.go**, pastikan juga fungsi harus di kelompokan sesuai dengan fungsi nya dan jika tidak ada pegelompokan nya silahkan buat baru file tersebut. Oke setelah berhasil ditambahkan silahkan lakukan `pull request` ke `main branch`. Kemudian nanti code tersebut akan di merge oleh administrator dan akan di buatkan tagging.