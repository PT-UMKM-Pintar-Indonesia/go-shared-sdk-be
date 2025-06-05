# Shared SDK

Repository ini berisi shared SDK yang akan digunakan untuk pegembangan aplikasi dan bersifat reusable tools.

# Install Dependency

- ###  Install Latest Version
```sh
go get github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1
```

- ### Install Specific Version
```sh
go get github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1@v0.0.1
```

## Code Of Conduct

Jika ingin menambahkan fungsi baru pastikan sesuai dengan tempat nya masing - masing, apakah itu **helpers**, **pkg**, **connection** dan sebagainya. Berikut adalah contoh saya menambahkan **helpers** `ToByte` berarti nanti fungsi tersebut akan masuk ke **helpers -> parser.helper.go**, pastikan juga fungsi harus di kelompokan sesuai dengan fungsi nya dan jika tidak ada pegelompokan nya silahkan buat baru file tersebut. Oke setelah berhasil ditambahkan silahkan lakukan `pull request` ke `main branch`. Kemudian nanti code tersebut akan di merge oleh administrator dan akan di buatkan tagging.