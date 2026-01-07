package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// --- wallet ---
type Wallet struct {
	MSPID   string  `json:"MSPID"`
	Balance float64 `json:"Balance"`
}
// ... Struktur Checkpoint (untuk data perjalanan) ...
type Checkpoint struct {
	Timestamp   string  `json:"Timestamp"`
	Location    string  `json:"Location"`
	Coordinates string  `json:"Coordinates"`
	Activity    string  `json:"Activity"`
	Actor       string  `json:"Actor"`
}

// ... Struktur Utama ...
type KopiBatch struct {
	BatchID            string       `json:"BatchID"`
	Status             string       `json:"Status"` 
	Petani             string       `json:"Petani"`
	KebunGeo           string       `json:"KebunGeo"`
	NamaLokasi         string       `json:"NamaLokasi"`
	JenisKopi          string       `json:"JenisKopi"`
	TanggalPanen       string       `json:"TanggalPanen"`
	
	// Logistik Lokal
	SupirTruk          string       `json:"SupirTruk"`
	PlatNomor          string       `json:"PlatNomor"`
	SuhuTrukLokal      float64      `json:"SuhuTrukLokal"`
	BiayaLogistikLokal float64      `json:"BiayaLogistikLokal"`
	
	// Organisasi/Koperasi
	LokasiGudang       string       `json:"LokasiGudang"`
	MetodePascapanen   string       `json:"MetodePascapanen"`
	TglTibaGudang      string       `json:"TglTibaGudang"`
	SkorCupping        int          `json:"SkorCupping"` 
	SkorCuppingFinal   int          `json:"SkorCuppingFinal"`
	TesResidu          string       `json:"TesResidu"`
	BeratBersih        float64      `json:"BeratBersih"`
	TglProsesQC        string       `json:"TglProsesQC"`
	
	// Ekspor
	DokumenEkspor      string       `json:"DokumenEkspor"`
	StatusIzin         string       `json:"StatusIzin"`
	TglIzinEkspor      string       `json:"TglIzinEkspor"`
	NamaKapal          string       `json:"NamaKapal"`
	NoKontainer        string       `json:"NoKontainer"`
	SuhuKontainer      float64      `json:"SuhuKontainer"`
	TglBerangkat       string       `json:"TglBerangkat"`
	TglTerima          string       `json:"TglTerima"`
	Importir           string       `json:"Importir"`
	
	// Keuangan
	NilaiKontrakEkspor float64      `json:"NilaiKontrakEkspor"`
	UangMuka           float64      `json:"UangMuka"`
	PotonganDenda      float64      `json:"PotonganDenda"`
	BonusKualitas      float64      `json:"BonusKualitas"`
	SisaTagihan        float64      `json:"SisaTagihan"`
	TotalDibayar       float64      `json:"TotalDibayar"`
	FinalPayout        float64      `json:"FinalPayout"`
	StatusPembayaran   string       `json:"StatusPembayaran"`
	CatatanMasalah     string       `json:"CatatanMasalah"`

	JourneyHistory     []Checkpoint `json:"JourneyHistory"`
}


// --- HELPER TIME & MSP ---
func getTxTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil { return "", err }
	return time.Unix(txTimestamp.Seconds, 0).UTC().Format("2006-01-02 15:04:05 MST"), nil
}

func checkMSPID(ctx contractapi.TransactionContextInterface, expectedMSPID string) error {
	clientMSPIDStr, _ := ctx.GetClientIdentity().GetMSPID()
	if clientMSPIDStr != expectedMSPID { return fmt.Errorf("AKSES DITOLAK: Fungsi ini hanya untuk %s (Anda: %s)", expectedMSPID, clientMSPIDStr) }
	return nil
}

// --- WALLET LOGIC ---
func (s *SmartContract) InitWallet(ctx contractapi.TransactionContextInterface) error {
	walletImportir := Wallet{MSPID: "ImportirMSP", Balance: 10000000000} 
	walletGudang := Wallet{MSPID: "KoperasiMSP", Balance: 5000000000} 
	walletPetani := Wallet{MSPID: "PetaniMSP", Balance: 0}

	importirJSON, _ := json.Marshal(walletImportir)
	gudangJSON, _ := json.Marshal(walletGudang)
	petaniJSON, _ := json.Marshal(walletPetani)

	ctx.GetStub().PutState("WALLET_ImportirMSP", importirJSON)
	ctx.GetStub().PutState("WALLET_KoperasiMSP", gudangJSON)
	ctx.GetStub().PutState("WALLET_PetaniMSP", petaniJSON)
	return nil
}

func (s *SmartContract) GetWalletBalance(ctx contractapi.TransactionContextInterface, mspId string) (*Wallet, error) {
	walletJSON, err := ctx.GetStub().GetState("WALLET_" + mspId)
	if err != nil { return nil, err }
	if walletJSON == nil { return nil, fmt.Errorf("wallet %s belum dibuat. Jalankan InitWallet dulu", mspId) }
	var wallet Wallet
	json.Unmarshal(walletJSON, &wallet)
	return &wallet, nil
}

// transfer func
func (s *SmartContract) transferFunds(ctx contractapi.TransactionContextInterface, fromMSP string, toMSP string, amount float64) error {
	senderBytes, _ := ctx.GetStub().GetState("WALLET_" + fromMSP)
	if senderBytes == nil { return fmt.Errorf("Wallet pengirim %s tidak ditemukan", fromMSP) }
	var sender Wallet
	json.Unmarshal(senderBytes, &sender)

	receiverBytes, _ := ctx.GetStub().GetState("WALLET_" + toMSP)
	if receiverBytes == nil { return fmt.Errorf("Wallet penerima %s tidak ditemukan", toMSP) }
	var receiver Wallet
	json.Unmarshal(receiverBytes, &receiver)

	if sender.Balance < amount && fromMSP != "KoperasiMSP" { 
		return fmt.Errorf("SALDO TIDAK CUKUP! %s punya %.2f, butuh %.2f", fromMSP, sender.Balance, amount)
	}

	sender.Balance -= amount
	receiver.Balance += amount

	senderUpd, _ := json.Marshal(sender)
	receiverUpd, _ := json.Marshal(receiver)
	ctx.GetStub().PutState("WALLET_" + fromMSP, senderUpd)
	ctx.GetStub().PutState("WALLET_" + toMSP, receiverUpd)

    type BankTransferEvent struct {
        From   string  `json:"From"`
        To     string  `json:"To"`
        Amount float64 `json:"Amount"`
    }
    eventPayload := BankTransferEvent{From: fromMSP, To: toMSP, Amount: amount}
    eventJSON, _ := json.Marshal(eventPayload)
    ctx.GetStub().SetEvent("BankTransfer", eventJSON) 

	return nil
}

// --- SUPPLY CHAIN ---

// 1. CreateBatch (DP 5% ke Petani)
func (s *SmartContract) CreateBatch(ctx contractapi.TransactionContextInterface,
	batchID string, petani string, geo string, namaLokasi string, jenis string, nilaiKontrak float64) error {

	if err := checkMSPID(ctx, "PetaniMSP"); err != nil { return err }

	exists, err := s.BatchExists(ctx, batchID)
	if err != nil { return err }
	if exists { return fmt.Errorf("batch %s sudah ada", batchID) }

	timestamp, _ := getTxTimestamp(ctx)
	// DP BARU: 5%
	dpAmount := nilaiKontrak * 0.05 

	// Bayar DP (Importir -> Petani)
	err = s.transferFunds(ctx, "ImportirMSP", "PetaniMSP", dpAmount)
	if err != nil { return fmt.Errorf("GAGAL BAYAR DP: %s", err.Error()) }

	firstPoint := Checkpoint{
		Timestamp: timestamp, Location: namaLokasi, Coordinates: geo, Activity: "Panen & Registrasi Awal", Actor: "Petani",
	}

	batch := KopiBatch{
		BatchID:            batchID,
		Status:             "PANEN_PETANI", 
		Petani:             petani,
		KebunGeo:           geo,
		NamaLokasi:         namaLokasi,
		JenisKopi:          jenis,
		TanggalPanen:       timestamp,
		NilaiKontrakEkspor: nilaiKontrak,
		UangMuka:           dpAmount,
		TotalDibayar:       dpAmount,
		SisaTagihan:        nilaiKontrak - dpAmount, // Sisa tagihan 95%
		StatusPembayaran:   "DP_5%_CAIR_KE_PETANI",
		CatatanMasalah:     "-",
		JourneyHistory:     []Checkpoint{firstPoint},
	}

	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 2. Transport Lokal
func (s *SmartContract) TransportToWarehouse(ctx contractapi.TransactionContextInterface,
	batchID string, supir string, plat string, suhu float64, currentGeo string) error {

	if err := checkMSPID(ctx, "LogistikMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "PANEN_PETANI" { 
		return fmt.Errorf("ALUR SALAH: Status saat ini '%s', harusnya 'PANEN_PETANI'", batch.Status) 
	}

	timestamp, _ := getTxTimestamp(ctx)
	batch.Status = "MENUJU_GUDANG" 
	batch.SupirTruk = supir
	batch.PlatNomor = plat
	batch.SuhuTrukLokal = suhu

	// Denda Suhu Lokal
	if suhu > 30.0 {
		batch.CatatanMasalah = "DENDA: Suhu Truk Lokal Panas"
		batch.PotonganDenda += 500000 
	}

	newPoint := Checkpoint{
		Timestamp: timestamp, Location: "Pick Up Point (Kebun)", Coordinates: currentGeo, Activity: fmt.Sprintf("Diangkut Truk %s (Suhu: %.1f°C)", plat, suhu), Actor: "Logistik Lokal",
	}
	batch.JourneyHistory = append(batch.JourneyHistory, newPoint)
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// Fitur Tambahan: AddCheckpoint
func (s *SmartContract) AddCheckpoint(ctx contractapi.TransactionContextInterface,
	batchID string, locationName string, geo string, activity string) error {

	if err := checkMSPID(ctx, "LogistikMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "MENUJU_GUDANG" && batch.Status != "DALAM_PERJALANAN_EKSPOR" {
		return fmt.Errorf("tidak bisa update posisi: barang harus sedang dalam perjalanan")
	}

	timestamp, _ := getTxTimestamp(ctx)
	newPoint := Checkpoint{
		Timestamp: timestamp, Location: locationName, Coordinates: geo, Activity: activity, Actor: "Logistik/Transporter",
	}
	batch.JourneyHistory = append(batch.JourneyHistory, newPoint)
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 3. Receive Gudang (Organisasi)
func (s *SmartContract) ReceiveAtWarehouse(ctx contractapi.TransactionContextInterface,
	batchID string, namaGudang string, kota string, geo string) error {

	if err := checkMSPID(ctx, "KoperasiMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "MENUJU_GUDANG" { 
		return fmt.Errorf("ALUR SALAH: Barang belum dikirim oleh logistik (Status: %s)", batch.Status) 
	}

	timestamp, _ := getTxTimestamp(ctx)
	batch.Status = "DITERIMA_GUDANG_KOPERASI" 
	batch.LokasiGudang = fmt.Sprintf("%s, %s", namaGudang, kota)
	batch.TglTibaGudang = timestamp

	newPoint := Checkpoint{
		Timestamp: timestamp, Location: namaGudang, Coordinates: geo, Activity: "Inbound Gudang (Barang Masuk)", Actor: "Koperasi",
	}
	batch.JourneyHistory = append(batch.JourneyHistory, newPoint)
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 3b. QC Process (Organisasi/Koperasi)
func (s *SmartContract) ProcessAndQC(ctx contractapi.TransactionContextInterface,
	batchID string, metode string, skor int, residu string, berat float64) error {

	if err := checkMSPID(ctx, "KoperasiMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "DITERIMA_GUDANG_KOPERASI" { 
		return fmt.Errorf("ALUR SALAH: Barang belum diterima di gudang secara resmi (Status: %s)", batch.Status) 
	}

	batch.MetodePascapanen = metode
	batch.SkorCupping = skor
	batch.TesResidu = residu
	batch.BeratBersih = berat
	batch.TglProsesQC, _ = getTxTimestamp(ctx)

	if skor < 70 || residu == "TERKONTAMINASI" {
		batch.Status = "GAGAL_QC"
		batch.CatatanMasalah = "GAGAL EKSPOR: Kualitas Buruk / Residu"
		batch.SisaTagihan = 0 
		
		// REFUND DP DITANGGUNG ORGANISASI (KoperasiMSP)
		err := s.transferFunds(ctx, "KoperasiMSP", "ImportirMSP", batch.UangMuka)
		if err != nil {
			return fmt.Errorf("GAGAL REFUND DARI ORGANISASI: %s", err.Error())
		}
		
		batch.StatusPembayaran = "REFUND_DP_SUKSES_DITANGGUNG_ORG"
	} else {
		batch.Status = "SIAP_EKSPOR"
		if skor > 85 {
			batch.BonusKualitas = batch.NilaiKontrakEkspor * 0.10
		}
	}
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 4. Regulator Approve
func (s *SmartContract) ApproveExport(ctx contractapi.TransactionContextInterface,
	batchID string, noDokumen string, keputusan string) error {

	if err := checkMSPID(ctx, "RegulatorMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "SIAP_EKSPOR" {
		if batch.Status == "GAGAL_QC" { return fmt.Errorf("barang Gagal qc, tidak bisa diajukan izin") }
		return fmt.Errorf("ALUR SALAH: Barang belum melewati QC atau belum siap ekspor (Status: %s)", batch.Status)
	}

	batch.DokumenEkspor = noDokumen
	batch.StatusIzin = keputusan
	batch.TglIzinEkspor, _ = getTxTimestamp(ctx)

	if keputusan == "APPROVED" {
		batch.Status = "LOLOS_INSPEKSI_REGULATOR"
	} else {
		batch.Status = "DITOLAK_REGULATOR"
		
		// REFUND DP DITANGGUNG ORGANISASI (KoperasiMSP)
		err := s.transferFunds(ctx, "KoperasiMSP", "ImportirMSP", batch.UangMuka)
		if err != nil {
			return fmt.Errorf("GAGAL REFUND DARI ORGANISASI: %s", err.Error())
		}
		
		batch.SisaTagihan = 0
		batch.StatusPembayaran = "REFUND_DITOLAK_REGULATOR_ORG_TANGGUNG"
	}
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 5. Shipping Start (Logistik Pihak Ketiga)
func (s *SmartContract) StartExportShipment(ctx contractapi.TransactionContextInterface,
	batchID string, kapal string, kontainer string, suhu float64, originGeo string) error {

	if err := checkMSPID(ctx, "LogistikMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "LOLOS_INSPEKSI_REGULATOR" { 
		return fmt.Errorf("ALUR SALAH: Belum ada izin ekspor dari Regulator (Status: %s)", batch.Status) 
	}

	timestamp, _ := getTxTimestamp(ctx)
	batch.Status = "DALAM_PERJALANAN_EKSPOR"
	batch.NamaKapal = kapal
	batch.NoKontainer = kontainer
	batch.SuhuKontainer = suhu
	batch.TglBerangkat = timestamp

	// Denda Logistik Ekspor (Ditanggung Organisasi/Petani)
	if suhu > 25.0 {
		msg := " DENDA: Kontainer Overheat"
		if batch.CatatanMasalah == "-" { batch.CatatanMasalah = msg } else { batch.CatatanMasalah += msg }
		dendaPanas := batch.NilaiKontrakEkspor * 0.05
		batch.PotonganDenda += dendaPanas
	}
	newPoint := Checkpoint{
		Timestamp: timestamp, Location: "Port of Origin (Loading)", Coordinates: originGeo, Activity: "Loading Kapal Ekspor", Actor: "Freight Forwarder",
	}
	batch.JourneyHistory = append(batch.JourneyHistory, newPoint)
	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}

// 6. Importir Confirm (QC Final & Settlement)
func (s *SmartContract) ConfirmImport(ctx contractapi.TransactionContextInterface,
	batchID string, namaImportir string, destGeo string, skorAkhir int) error { 

	if err := checkMSPID(ctx, "ImportirMSP"); err != nil { return err }
	batch, err := s.ReadBatch(ctx, batchID)
	if err != nil { return err }

	if batch.Status != "DALAM_PERJALANAN_EKSPOR" { 
		return fmt.Errorf("ALUR SALAH: Barang belum sampai atau status salah") 
	}

	timestamp, _ := getTxTimestamp(ctx)
	
	batch.SkorCuppingFinal = skorAkhir 
	
	// === LOGIKA PENERIMAAN / PENOLAKAN AKHIR ===
	const ACCEPTANCE_MIN_SCORE = 80
	const ORG_COMMISSION_RATE = 0.20 // 20% Komisi Organisasi
	const FARMER_SHARE_RATE = 0.80 // 80% Porsi Petani
    // Denda Persentase
    const PENALTY_RATE_PER_POINT = 0.01 // 1% dari nilai kontrak per poin

	if skorAkhir < ACCEPTANCE_MIN_SCORE {
		// Skenario 1: REJECT (Kualitas Buruk di Tujuan)
		batch.Status = "DITOLAK_IMPORTIR"
		batch.Importir = namaImportir
		batch.TglTerima = timestamp
		
		msg := fmt.Sprintf("REJECTED: Skor Landed Quality hanya %d (Min %d).", skorAkhir, ACCEPTANCE_MIN_SCORE)
		if batch.CatatanMasalah == "-" { batch.CatatanMasalah = msg } else { batch.CatatanMasalah += " | " + msg }

		// Keuangan: Refund DP DITANGGUNG ORGANISASI
		err := s.transferFunds(ctx, "KoperasiMSP", "ImportirMSP", batch.UangMuka)
		if err != nil { return fmt.Errorf("GAGAL REFUND DARI ORGANISASI (REJECT): %s", err.Error()) }

		batch.SisaTagihan = 0
		batch.StatusPembayaran = "REFUND_DP_DITANGGUNG_ORG_REJECT"
		
	} else {
		// Skenario 2: Diterima (ACCEPTANCE & PENALTY)
		batch.Status = "DITERIMA_IMPORTIR"
		batch.Importir = namaImportir
		batch.TglTerima = timestamp

		// Hitung Denda Penurunan Mutu (Skor Awal QC Gudang vs Skor Landed Importir)
		selisihSkor := batch.SkorCupping - skorAkhir
		if selisihSkor > 0 {
			dendaMutu := batch.NilaiKontrakEkspor * PENALTY_RATE_PER_POINT * float64(selisihSkor) 

			batch.PotonganDenda += dendaMutu
			batch.CatatanMasalah += fmt.Sprintf(" | Penurunan Mutu %d Poin (Potongan %.2f%%)", selisihSkor, PENALTY_RATE_PER_POINT * float64(selisihSkor) * 100) 
		}

		// 1. Hitung Net Transfer
		finalTransfer := batch.SisaTagihan + batch.BonusKualitas - batch.PotonganDenda
		if finalTransfer < 0 { finalTransfer = 0 }
        
        // 2. Pembagian Komisi 80/20
        orgShare := finalTransfer * ORG_COMMISSION_RATE
        farmerShare := finalTransfer * FARMER_SHARE_RATE

        // 3. Transfer Komisi ke ORGANISASI (KoperasiMSP)
        s.transferFunds(ctx, "ImportirMSP", "KoperasiMSP", orgShare)

        // 4. Transfer Sisa Pembayaran ke PETANI (PetaniMSP)
        s.transferFunds(ctx, "ImportirMSP", "PetaniMSP", farmerShare)

		batch.SisaTagihan = 0 
		batch.TotalDibayar = batch.UangMuka + finalTransfer
		batch.FinalPayout = finalTransfer 
		batch.StatusPembayaran = "LUNAS_FINAL_SETTLEMENT"
	}

	// Checkpoint Terakhir
	newPoint := Checkpoint{
		Timestamp: timestamp, Location: "Gudang Importir", Coordinates: destGeo, 
		Activity: fmt.Sprintf("Verifikasi Landed Quality (Skor: %d)", skorAkhir), Actor: namaImportir,
	}
	batch.JourneyHistory = append(batch.JourneyHistory, newPoint)

	batchJSON, _ := json.Marshal(batch)
	return ctx.GetStub().PutState(batchID, batchJSON)
}


// -------------- HELPERS ------------------
// Helpers Read
func (s *SmartContract) ReadBatch(ctx contractapi.TransactionContextInterface, batchID string) (*KopiBatch, error) {
	batchJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil { return nil, err }
	if batchJSON == nil { return nil, fmt.Errorf("batch %s tidak ditemukan", batchID) }
	var batch KopiBatch
	err = json.Unmarshal(batchJSON, &batch)
	return &batch, err
}

func (s *SmartContract) BatchExists(ctx contractapi.TransactionContextInterface, batchID string) (bool, error) {
	batchJSON, err := ctx.GetStub().GetState(batchID)
	return batchJSON != nil, err
}

func (s *SmartContract) GetAllBatches(ctx contractapi.TransactionContextInterface) ([]*KopiBatch, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil { return nil, err }
	defer resultsIterator.Close()

	var batches []*KopiBatch
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil { return nil, err }
		if len(queryResponse.Key) >= 7 && queryResponse.Key[:7] == "WALLET_" {
			continue 
		}
		var batch KopiBatch
		err = json.Unmarshal(queryResponse.Value, &batch)
		if err != nil {
			continue 
		}
		
		if batch.BatchID != "" {
			batches = append(batches, &batch)
		}
	}
	return batches, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil { log.Panicf("Error creating chaincode: %v", err) }
	if err := chaincode.Start(); err != nil { log.Panicf("Error starting chaincode: %v", err) }
}