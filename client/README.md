# Outbound Client Layer (Gateways)

Folder ini berfungsi sebagai **pintu gerbang keluar (Outbound Gateway/Adapter)** untuk melakukan pemanggilan ke sistem eksternal atau pihak ketiga (Third-Party APIs), seperti SMS gateway, Payment gateway, Microservice lain, dsb.

Arsitektur di folder ini menerapkan **Dependency Inversion Principle (DIP)** agar *layer bisnis* (Service) tidak terikat langsung secara keras (*hard-coupled*) dengan detail teknis API luar.

---

## Struktur Folder

```text
client/
├── restclient/                 # Implementasi konkret dari interface client
│   └── httpbin_restclient.go   # Implementasi konkret menggunakan HTTP/REST
├── README.md                   # Dokumentasi ini
└── httpbin_client.go           # Interface (Kontrak) netral
```

---

## Cara Penggunaan (Pola Clean Architecture)

### 1. Definisikan Interface Umum di folder `client`
Buatlah kontrak yang netral (bebas dari nama penyedia/vendor tertentu).

Contoh `client/payment_client.go`:
```go
package client

import "context"

type PaymentClient interface {
	SendPayment(ctx context.Context, invoiceID string, amount int64) (string, error)
}
```

### 2. Implementasikan secara Teknis di subfolder `restclient`
Tulis logika HTTP request sesungguhnya di sini. Anda bisa membuat banyak implementasi untuk satu interface yang sama.

Contoh `client/restclient/midtrans_client.go`:
```go
package restclient

import (
	"context"
	"evasbr/mclamg/client"
	"evasbr/mclamg/common"
)

type MidtransClient struct {}

func NewMidtransClient() client.PaymentClient {
	return &MidtransClient{}
}

func (m *MidtransClient) SendPayment(ctx context.Context, invoiceID string, amount int64) (string, error) {
	// Integrasi teknis spesifik ke Midtrans menggunakan common.ClientComponent
	// ...
	return "midtrans_token_123", nil
}
```

### 3. Gunakan Interface di Service (Dependency Injection)
Layer *Service* tidak boleh tahu-menahu tentang "Midtrans" atau "Xendit". Service hanya tahu interface `client.PaymentClient`.

Contoh `service/impl/order_service_impl.go`:
```go
type orderServiceImpl struct {
	paymentClient client.PaymentClient // Inject interface, bukan struct konkret
}

func (s *orderServiceImpl) ProcessOrder(ctx context.Context, orderID string) {
	// ...
	token, err := s.paymentClient.SendPayment(ctx, orderID, 150000)
	// ...
}
```

### 4. Hubungkan di `main.go`
Saat aplikasi melakukan inisialisasi, Anda tinggal meng-inject implementasi mana yang ingin digunakan.

```go
// Jika ingin menggunakan Midtrans:
paymentRestClient := restclient.NewMidtransClient()

// Suntikkan ke Service:
orderService := service.NewOrderServiceImpl(paymentRestClient)
```

---

## Keuntungan Pola Ini

1. **Fleksibilitas Pergantian Vendor (API):**
   Jika di kemudian hari sistem ingin bermigrasi dari **Midtrans** ke **Xendit**, Anda hanya perlu membuat berkas implementasi `xendit_client.go` baru di `restclient` dan merubah satu baris inisialisasi di `main.go`. Kode di dalam `Service` Anda **sama sekali tidak perlu diubah atau dites ulang**.
2. **Kemudahan Unit Testing (Mocking):**
   Saat membuat unit test untuk *Service*, Anda tidak ingin menembak server pihak ketiga asli (karena lambat dan memerlukan internet). Anda dapat membuat *Mock Client* tiruan yang mengimplementasikan interface `client.PaymentClient` dan meng-inject-nya ke dalam Service secara offline.
