# ☕ Chaincode Supply Chain Ekspor Kopi Indonesia

![Hyperledger Fabric](https://img.shields.io/badge/Hyperledger%20Fabric-2.2.3-blue)
![Go](https://img.shields.io/badge/Go-1.x-00ADD8)
![Node.js](https://img.shields.io/badge/Node.js-Express-green)
![License](https://img.shields.io/badge/License-MIT-yellow)

Sistem blockchain berbasis **Hyperledger Fabric** untuk traceability dan transparansi rantai pasok ekspor kopi Indonesia. Sistem ini melibatkan **5 organisasi** yang berperan dalam proses dari panen hingga ekspor ke importir internasional.

## 📋 Daftar Isi

- [Fitur Utama](#-fitur-utama)
- [Arsitektur Sistem](#-arsitektur-sistem)
- [Organisasi & Peran](#-organisasi--peran)
- [Alur Proses (Supply Chain Flow)](#-alur-proses-supply-chain-flow)
- [Struktur Data](#-struktur-data)
- [Prasyarat](#-prasyarat)
- [Instalasi](#-instalasi)
- [Menjalankan Jaringan](#-menjalankan-jaringan)
- [API Endpoints](#-api-endpoints)
- [Fungsi Smart Contract](#-fungsi-smart-contract)
- [Mekanisme Keuangan](#-mekanisme-keuangan)
- [Contoh Penggunaan](#-contoh-penggunaan)
- [Troubleshooting](#-troubleshooting)

---

## 🚀 Fitur Utama

- ✅ **Full Traceability** - Lacak perjalanan kopi dari kebun petani hingga gudang importir
- ✅ **Multi-Organization** - 5 organisasi dengan peran dan akses berbeda
- ✅ **Smart Contract** - Otomatisasi proses bisnis dengan chaincode Go
- ✅ **Digital Wallet** - Sistem pembayaran terintegrasi di blockchain
- ✅ **Quality Control** - Pencatatan skor cupping dan tes residu
- ✅ **Geo-Tracking** - Checkpoint dengan koordinat GPS
- ✅ **Endorsement Policy** - Minimal 3 dari 5 organisasi untuk validasi
- ✅ **Event-Driven** - Notifikasi real-time untuk transfer bank
- ✅ **Penalty System** - Denda otomatis untuk pelanggaran suhu dan kualitas
- ✅ **Commission Split** - Pembagian hasil 80% petani, 20% koperasi

---

## 🏗 Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Frontend / Client App                          │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         REST API Server (Express.js)                     │
│                              Port: 3001                                  │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Hyperledger Fabric Network                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ Petani  │  │Logistik │  │Koperasi │  │Regulator│  │Importir │        │
│  │  Peer0  │  │  Peer0  │  │  Peer0  │  │  Peer0  │  │  Peer0  │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │            │              │
│       └────────────┴────────────┴────────────┴────────────┘              │
│                              mychannel                                   │
│                                  │                                       │
│                      ┌───────────┴───────────┐                          │
│                      │   Chaincode: "kopi"   │                          │
│                      │       (Golang)        │                          │
│                      └───────────────────────┘                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 👥 Organisasi & Peran

| No | Organisasi | MSP ID | Domain | Peran |
|----|------------|--------|--------|-------|
| 1 | **Petani** | `PetaniMSP` | petani.co.id | Mendaftarkan panen baru, menerima pembayaran |
| 2 | **Logistik** | `LogistikMSP` | logistik.co.id | Transportasi lokal & ekspor internasional |
| 3 | **Koperasi** | `KoperasiMSP` | koperasi.co.id | Gudang, QC, pascapanen, menanggung risiko |
| 4 | **Regulator** | `RegulatorMSP` | regulator.co.id | Inspeksi & persetujuan dokumen ekspor |
| 5 | **Importir** | `ImportirMSP` | importir.co.id | Pembeli akhir, verifikasi kualitas tiba |

---

## 📊 Alur Proses (Supply Chain Flow)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   PETANI     │     │   LOGISTIK   │     │   KOPERASI   │     │   LOGISTIK   │
│              │     │    LOKAL     │     │   (GUDANG)   │     │   EKSPOR     │
│  CreateBatch │────▶│  Transport   │────▶│  Receive +   │────▶│  Start       │
│   (Panen)    │     │  ToWarehouse │     │  ProcessQC   │     │  Shipment    │
│              │     │              │     │              │     │              │
│ DP 5% Cair   │     │ Checkpoint   │     │ Skor Cupping │     │ Loading      │
│              │     │ GPS Track    │     │ Tes Residu   │     │ Kontainer    │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                │                      │
                     ┌──────────────┐           │                      │
                     │  REGULATOR   │◀──────────┘                      │
                     │              │                                  │
                     │ Approve/     │                                  │
                     │ Reject       │                                  │
                     │ Export       │                                  │
                     └──────────────┘                                  │
                                                                       ▼
                                          ┌──────────────────────────────────┐
                                          │          IMPORTIR                │
                                          │                                  │
                                          │  ConfirmImport                   │
                                          │  (QC Final + Settlement)         │
                                          │                                  │
                                          │  ✓ Diterima → Pelunasan 95%     │
                                          │  ✗ Ditolak  → Refund DP         │
                                          └──────────────────────────────────┘
```

### Status Batch

| Status | Deskripsi |
|--------|-----------|
| `PANEN_PETANI` | Batch baru didaftarkan petani |
| `MENUJU_GUDANG` | Dalam transportasi ke gudang koperasi |
| `DITERIMA_GUDANG_KOPERASI` | Tiba di gudang |
| `GAGAL_QC` | Kualitas tidak memenuhi standar |
| `SIAP_EKSPOR` | Lolos QC, siap diajukan ke regulator |
| `LOLOS_INSPEKSI_REGULATOR` | Dokumen ekspor disetujui |
| `DITOLAK_REGULATOR` | Ditolak regulator |
| `DALAM_PERJALANAN_EKSPOR` | Dalam pengiriman ke importir |
| `DITERIMA_IMPORTIR` | Diterima dengan baik di tujuan |
| `DITOLAK_IMPORTIR` | Ditolak karena kualitas menurun |

---

## 📦 Struktur Data

### KopiBatch (Struktur Utama)

```go
type KopiBatch struct {
    BatchID            string       // ID unik batch
    Status             string       // Status saat ini
    Petani             string       // Nama petani
    KebunGeo           string       // Koordinat GPS kebun
    NamaLokasi         string       // Nama lokasi/desa
    JenisKopi          string       // Arabica/Robusta
    TanggalPanen       string       // Timestamp panen
    
    // Logistik Lokal
    SupirTruk          string       // Nama supir
    PlatNomor          string       // Plat kendaraan
    SuhuTrukLokal      float64      // Suhu saat transportasi (°C)
    BiayaLogistikLokal float64      // Biaya transportasi
    
    // Organisasi/Koperasi
    LokasiGudang       string       // Nama & kota gudang
    MetodePascapanen   string       // Honey/Washed/Natural
    TglTibaGudang      string       // Timestamp tiba
    SkorCupping        int          // Skor QC awal (0-100)
    SkorCuppingFinal   int          // Skor QC di importir
    TesResidu          string       // BERSIH/TERKONTAMINASI
    BeratBersih        float64      // Berat setelah sortir (kg)
    TglProsesQC        string       // Timestamp QC
    
    // Ekspor
    DokumenEkspor      string       // Nomor dokumen
    StatusIzin         string       // APPROVED/REJECTED
    TglIzinEkspor      string       // Timestamp izin
    NamaKapal          string       // Nama kapal/vessel
    NoKontainer        string       // Nomor kontainer
    SuhuKontainer      float64      // Suhu kontainer (°C)
    TglBerangkat       string       // Timestamp keberangkatan
    TglTerima          string       // Timestamp diterima importir
    Importir           string       // Nama perusahaan importir
    
    // Keuangan
    NilaiKontrakEkspor float64      // Total nilai kontrak (IDR)
    UangMuka           float64      // DP 5%
    PotonganDenda      float64      // Total denda
    BonusKualitas      float64      // Bonus skor > 85
    SisaTagihan        float64      // Sisa yang harus dibayar
    TotalDibayar       float64      // Total yang sudah dibayar
    FinalPayout        float64      // Transfer terakhir
    StatusPembayaran   string       // Status pembayaran
    CatatanMasalah     string       // Catatan denda/masalah

    JourneyHistory     []Checkpoint // Riwayat perjalanan
}
```

### Checkpoint (Tracking Perjalanan)

```go
type Checkpoint struct {
    Timestamp   string  // Waktu checkpoint
    Location    string  // Nama lokasi
    Coordinates string  // Koordinat GPS
    Activity    string  // Aktivitas yang dilakukan
    Actor       string  // Pelaku/organisasi
}
```

### Wallet (Dompet Digital)

```go
type Wallet struct {
    MSPID   string  // ID organisasi
    Balance float64 // Saldo (IDR)
}
```

---

## ⚙️ Prasyarat

Pastikan sistem Anda sudah terinstall:

- **Docker** & **Docker Compose** (untuk Fabric network)
- **Node.js** v14+ (untuk REST API)
- **Go** 1.16+ (untuk chaincode development)
- **Fablo** (Hyperledger Fabric deployment tool)

```bash
# Install Fablo
npm install -g fablo

# Verifikasi instalasi
fablo version
```

---

## 🔧 Instalasi

### 1. Clone Repository

```bash
git clone <repository-url>
cd chaincode-supplychain-coffee-export
```

### 2. Install Dependencies

```bash
npm install
```

### 3. Konfigurasi Fablo

File `fablo-config.json` sudah dikonfigurasi dengan 5 organisasi. Review jika perlu.

---

## 🚀 Menjalankan Jaringan

### 1. Generate & Start Network

```bash
# Generate konfigurasi jaringan
fablo generate

# Start jaringan Fabric
fablo up
```

### 2. Enroll User untuk Setiap Organisasi

```bash
node enrollUp.js
```

Output yang diharapkan:
```
--- Processing Organization: Petani (PetaniMSP) ---
   ✅ Sukses Enroll Admin: admin-petani
   ✅ Sukses Enroll User (Registered): PetaniUser

--- Processing Organization: Logistik (LogistikMSP) ---
   ✅ Sukses Enroll Admin: admin-logistik
   ✅ Sukses Enroll User (Registered): LogistikUser
   
... (dan seterusnya untuk 5 organisasi)

 SELESAI. Semua 5 Organisasi Siap!
```

### 3. Jalankan REST API Server

```bash
npm start
# atau
node server.js
```

Server berjalan di `http://localhost:3001`

### 4. Inisialisasi Wallet Blockchain

```bash
curl -X POST http://localhost:3001/api/init-wallet
```

Ini akan membuat:
- Wallet Importir: IDR 10.000.000.000
- Wallet Koperasi: IDR 5.000.000.000
- Wallet Petani: IDR 0

---

## 🔌 API Endpoints

### Supply Chain Operations

| Method | Endpoint | Organisasi | Deskripsi |
|--------|----------|------------|-----------|
| `POST` | `/api/create` | Petani | Buat batch baru (panen) |
| `POST` | `/api/transport-lokal` | Logistik | Mulai transportasi ke gudang |
| `POST` | `/api/add-checkpoint` | Logistik | Tambah checkpoint perjalanan |
| `POST` | `/api/receive-warehouse` | Koperasi | Terima barang di gudang |
| `POST` | `/api/process-qc` | Koperasi | Proses QC & pascapanen |
| `POST` | `/api/approve-export` | Regulator | Approve/reject dokumen ekspor |
| `POST` | `/api/start-shipment` | Logistik | Mulai pengiriman ekspor |
| `POST` | `/api/confirm-import` | Importir | Konfirmasi penerimaan & QC final |

### Wallet Operations

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `POST` | `/api/init-wallet` | Inisialisasi saldo awal |
| `GET` | `/api/wallet/:mspId` | Cek saldo organisasi |

### Query Operations

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `GET` | `/api/all-batches` | Ambil semua batch |
| `GET` | `/api/batch/:id` | Ambil detail batch |

---

## 📜 Fungsi Smart Contract

### Fungsi Utama

| Fungsi | MSP Required | Deskripsi |
|--------|--------------|-----------|
| `InitWallet()` | Any | Inisialisasi wallet semua organisasi |
| `CreateBatch()` | PetaniMSP | Daftarkan panen baru + DP 5% |
| `TransportToWarehouse()` | LogistikMSP | Transportasi lokal |
| `AddCheckpoint()` | LogistikMSP | Tambah titik tracking |
| `ReceiveAtWarehouse()` | KoperasiMSP | Terima di gudang |
| `ProcessAndQC()` | KoperasiMSP | Quality control |
| `ApproveExport()` | RegulatorMSP | Approve/reject ekspor |
| `StartExportShipment()` | LogistikMSP | Mulai shipping |
| `ConfirmImport()` | ImportirMSP | Konfirmasi & settlement |

### Fungsi Helper

| Fungsi | Deskripsi |
|--------|-----------|
| `ReadBatch()` | Baca detail batch |
| `BatchExists()` | Cek keberadaan batch |
| `GetAllBatches()` | Ambil semua batch |
| `GetWalletBalance()` | Cek saldo wallet |

---

## 💰 Mekanisme Keuangan

### Alur Pembayaran

```
┌─────────────────────────────────────────────────────────────────┐
│                      ALUR PEMBAYARAN                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. CREATE BATCH                                                │
│     └── Importir → Petani: DP 5% dari nilai kontrak            │
│                                                                 │
│  2. CONFIRM IMPORT (Sukses)                                     │
│     └── Importir → Koperasi: 20% dari sisa (95%)               │
│     └── Importir → Petani: 80% dari sisa (95%)                 │
│                                                                 │
│  3. GAGAL QC / DITOLAK REGULATOR / DITOLAK IMPORTIR            │
│     └── Koperasi → Importir: Refund DP 5%                      │
│         (Koperasi menanggung risiko)                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Sistem Denda

| Kondisi | Denda |
|---------|-------|
| Suhu truk lokal > 30°C | IDR 500.000 |
| Suhu kontainer ekspor > 25°C | 5% dari nilai kontrak |
| Penurunan skor cupping | 1% per poin penurunan |

### Bonus Kualitas

| Kondisi | Bonus |
|---------|-------|
| Skor cupping > 85 | 10% dari nilai kontrak |

---

## 📝 Contoh Penggunaan

### 1. Petani Mendaftarkan Panen

```bash
curl -X POST http://localhost:3001/api/create \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "petani": "Pak Budi",
    "geo": "-6.914744, 107.609810",
    "namaLokasi": "Desa Pangalengan, Bandung",
    "jenis": "Arabica",
    "harga": 50000000
  }'
```

### 2. Logistik Mengangkut ke Gudang

```bash
curl -X POST http://localhost:3001/api/transport-lokal \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "supir": "Ahmad",
    "plat": "D 1234 ABC",
    "suhu": 28.5,
    "currentGeo": "-6.914744, 107.609810"
  }'
```

### 3. Koperasi Menerima di Gudang

```bash
curl -X POST http://localhost:3001/api/receive-warehouse \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "nama": "Gudang Koperasi Maju",
    "kota": "Bandung",
    "geo": "-6.905977, 107.613144"
  }'
```

### 4. Koperasi Melakukan QC

```bash
curl -X POST http://localhost:3001/api/process-qc \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "metode": "Honey Process",
    "skor": 87,
    "residu": "BERSIH",
    "berat": 950
  }'
```

### 5. Regulator Menyetujui Ekspor

```bash
curl -X POST http://localhost:3001/api/approve-export \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "dokumen": "EXP/2024/001234",
    "keputusan": "APPROVED"
  }'
```

### 6. Logistik Memulai Pengiriman Ekspor

```bash
curl -X POST http://localhost:3001/api/start-shipment \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "kapal": "MV Pacific Star",
    "kontainer": "MSKU1234567",
    "suhu": 22.0,
    "originGeo": "-6.102, 106.880"
  }'
```

### 7. Importir Konfirmasi Penerimaan

```bash
curl -X POST http://localhost:3001/api/confirm-import \
  -H "Content-Type: application/json" \
  -d '{
    "id": "BATCH001",
    "importir": "Tokyo Coffee Co.",
    "destGeo": "35.689, 139.691",
    "skorAkhir": 85
  }'
```

### 8. Cek Detail Batch

```bash
curl http://localhost:3001/api/batch/BATCH001
```

### 9. Cek Saldo Wallet

```bash
# Cek saldo petani
curl http://localhost:3001/api/wallet/PetaniMSP

# Cek saldo importir
curl http://localhost:3001/api/wallet/ImportirMSP

# Cek saldo koperasi
curl http://localhost:3001/api/wallet/KoperasiMSP
```

---

## 🛠 Troubleshooting

### Network Issues

```bash
# Restart jaringan
fablo down
fablo up

# Prune jaringan & buat ulang
fablo prune
fablo up
```

### Identity Not Found

```bash
# Hapus wallet & enroll ulang
rm -rf wallet/
node enrollUp.js
```

### Chaincode Error

```bash
# Upgrade chaincode
fablo chaincode upgrade kopi 1.1
```

### Connection Profile Not Found

Pastikan folder `fablo-target` sudah ada setelah menjalankan `fablo up`.

### Reset Wallet Blockchain

Jika ingin reset saldo wallet:
```bash
curl -X POST http://localhost:3001/api/init-wallet
```

---

## 📁 Struktur Folder

```
chaincode-supplychain-coffee-export/
├── enrollUp.js           # Script enroll user 5 organisasi
├── server.js             # REST API server
├── package.json          # Dependencies Node.js
├── fablo-config.json     # Konfigurasi Fablo (5 org)
├── fablo                 # Fablo executable
├── kopi/                 # Chaincode folder
│   ├── kopi.go           # Smart contract (Go)
│   ├── go.mod            # Go module
│   └── vendor/           # Go dependencies
├── wallet/               # Identitas user (generated)
└── fablo-target/         # Network config (generated)
```

---

## 📄 License

MIT License - Silakan gunakan dan modifikasi sesuai kebutuhan.

---

## 🤝 Kontribusi

1. Fork repository
2. Buat branch fitur (`git checkout -b feature/AmazingFeature`)
3. Commit perubahan (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buat Pull Request

---

## 📧 Kontak

Untuk pertanyaan atau dukungan, silakan buat issue di repository ini.

---

**Made with ❤️ for Indonesian Coffee Supply Chain**
