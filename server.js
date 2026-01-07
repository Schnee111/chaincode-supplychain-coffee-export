/* server.js - EDISI 5 ORGANISASI */
const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const fs = require('fs');

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const app = express();
app.use(cors());
app.use(bodyParser.json());
const PORT = 3001;

// --- Helper: Menentukan Profile Berdasarkan User ---
function getOrgProfile(identityName) {
    let orgName = '';
    let mspId = '';

    // Mapping User ke Organisasi & MSP
    switch (identityName) {
        case 'PetaniUser':
            orgName = 'petani';
            mspId = 'PetaniMSP';
            break;
        case 'LogistikUser':
            orgName = 'logistik';
            mspId = 'LogistikMSP';
            break;
        case 'KoperasiUser':
            orgName = 'koperasi';
            mspId = 'KoperasiMSP';
            break;
        case 'RegulatorUser':
            orgName = 'regulator';
            mspId = 'RegulatorMSP';
            break;
        case 'ImportirUser':
            orgName = 'importir';
            mspId = 'ImportirMSP';
            break;
        default:
            throw new Error(`User ${identityName} tidak dikenali!`);
    }

    const ccpPath = path.resolve(__dirname, `fablo-target/fabric-config/connection-profiles/connection-profile-${orgName}.json`);
    return { ccpPath, mspId };
}

// --- Helper: Connect Blockchain ---
async function connectToNetwork(identityName) {
    const { ccpPath } = getOrgProfile(identityName);
    
    if (!fs.existsSync(ccpPath)) {
        throw new Error(`CCP tidak ditemukan di: ${ccpPath}`);
    }

    const ccpContent = fs.readFileSync(ccpPath, 'utf8');
    const ccp = JSON.parse(ccpContent.replace(/http:\/\//g, 'https://'));

    const walletPath = path.join(process.cwd(), 'wallet');
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    
    const identity = await wallet.get(identityName);
    if (!identity) {
        throw new Error(`Identitas ${identityName} belum dibuat. Jalankan enrollUp.js!`);
    }

    const gateway = new Gateway();
    await gateway.connect(ccp, { 
        wallet, 
        identity: identityName, 
        discovery: { enabled: true, asLocalhost: true } 
    });
    
    const network = await gateway.getNetwork('mychannel');
    const contract = network.getContract('kopi');
    return { contract, gateway };
}

// --- FUNGSI LISTENER (SIMULASI BANK GATEWAY) ---
async function startBlockListener() {
    try {
        const { gateway, contract } = await connectToNetwork('PetaniUser');

        console.log("👂 Backend Listening for Chaincode Events...");

        // Tambahkan Listener
        await contract.addContractListener(async (event) => {
            if (event.eventName === 'BankTransfer') {
                const payload = JSON.parse(event.payload.toString());
                
                console.log("\n========================================");
                console.log("⚡ EVENT TRIGGERED: PERMINTAAN TRANSFER BANK");
                console.log("========================================");
                console.log(`📤 DARI REKENING : ${payload.From}`);
                console.log(`📥 KE REKENING   : ${payload.To}`);
                console.log(`💰 JUMLAH        : IDR ${payload.Amount.toLocaleString()}`);
                console.log("----------------------------------------");
                
                // SIMULASI PROSES BANK (Jeda 2 detik)
                console.log("🔄 Menghubungi API Bank (Simulasi)...");
                await new Promise(resolve => setTimeout(resolve, 2000));
                
                console.log("✅ TRANSFER BANK BERHASIL (Settlement Real-time)");
                console.log("========================================\n");
            }
        });

    } catch (error) {
        console.error(`Gagal setup listener: ${error}`);
    }
}

// Jalankan Listener saat server nyala
startBlockListener();

// ================= ROUTE API  =================

// 1. PETANI (Menggunakan PetaniUser)
app.post('/api/create', async (req, res) => {
    try {
        const { id, petani, geo, namaLokasi, jenis, harga } = req.body;
        const { contract, gateway } = await connectToNetwork('PetaniUser'); 
        
        console.log(`[CREATE] Batch: ${id} by Petani`);
        await contract.submitTransaction('CreateBatch', id, petani, geo, namaLokasi, jenis, harga.toString());
        
        await gateway.disconnect();
        res.json({ success: true, message: `Batch ${id} Panen Berhasil Dicatat!` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 2. LOGISTIK LOKAL (Menggunakan LogistikUser)
app.post('/api/transport-lokal', async (req, res) => {
    try {
        // Tambahkan currentGeo
        const { id, supir, plat, suhu, currentGeo } = req.body; 
        const { contract, gateway } = await connectToNetwork('LogistikUser');
        // Geo disini adalah lokasi PENJEMPUTAN (biasanya sama dengan kebun atau titik temu)
        await contract.submitTransaction('TransportToWarehouse', id, supir, plat, suhu.toString(), currentGeo || '-6.200, 106.816');
        await gateway.disconnect();
        res.json({ success: true, message: `Logistik Lokal Dimulai.` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 3. GUDANG (Menggunakan KoperasiUser)
app.post('/api/receive-warehouse', async (req, res) => {
    try {
        const { id, nama, kota, geo } = req.body;
        const { contract, gateway } = await connectToNetwork('KoperasiUser'); 
        await contract.submitTransaction('ReceiveAtWarehouse', id, nama, kota, geo);
        await gateway.disconnect();
        res.json({ success: true, message: `Diterima di Gudang.` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

app.post('/api/process-qc', async (req, res) => {
    try {
        const { id, metode, skor, residu, berat } = req.body;
        const { contract, gateway } = await connectToNetwork('KoperasiUser'); 
        await contract.submitTransaction('ProcessAndQC', id, metode, skor.toString(), residu, berat.toString());
        await gateway.disconnect();
        res.json({ success: true, message: `QC Selesai.` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 4. REGULATOR (Menggunakan RegulatorUser)
app.post('/api/approve-export', async (req, res) => {
    try {
        const { id, dokumen, keputusan } = req.body;
        const { contract, gateway } = await connectToNetwork('RegulatorUser'); 
        await contract.submitTransaction('ApproveExport', id, dokumen, keputusan);
        await gateway.disconnect();
        res.json({ success: true, message: `Keputusan Regulator: ${keputusan}` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 5. SHIPMENT (Menggunakan LogistikUser - Ekspedisi)
app.post('/api/start-shipment', async (req, res) => {
    try {
        const { id, kapal, kontainer, suhu, originGeo } = req.body;
        const { contract, gateway } = await connectToNetwork('LogistikUser'); 
        // Geo disini adalah Port of Origin
        await contract.submitTransaction('StartExportShipment', id, kapal, kontainer, suhu.toString(), originGeo || '-6.102, 106.880');
        await gateway.disconnect();
        res.json({ success: true, message: `Pengiriman Ekspor Dimulai.` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 6. IMPORTIR (Menggunakan ImportirUser)
app.post('/api/confirm-import', async (req, res) => {
    try {
        const { id, importir, destGeo, skorAkhir } = req.body; 
        const { contract, gateway } = await connectToNetwork('ImportirUser');
    
        const finalScore = skorAkhir ? skorAkhir.toString() : '80';
        await contract.submitTransaction('ConfirmImport', id, importir, destGeo || '35.689, 139.691', finalScore);
        
        await gateway.disconnect();
        res.json({ success: true, message: `Import Selesai. QC Final: ${finalScore}` });
    } catch (error) { 
        console.error(`Error saat Confirm Import: ${error.message}`);
        res.status(500).json({ success: false, error: error.message }); 
    }
});

// ADD CHECKPOINT
app.post('/api/add-checkpoint', async (req, res) => {
    try {
        const { id, location, geo, activity } = req.body;
        const { contract, gateway } = await connectToNetwork('LogistikUser');
        await contract.submitTransaction('AddCheckpoint', id, location, geo, activity);
        await gateway.disconnect();
        res.json({ success: true, message: `Checkpoint ${location} ditambahkan.` });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// --- FITUR WALLET ---

// 1. Inisialisasi Saldo (Jalankan sekali saja setelah reset network)
app.post('/api/init-wallet', async (req, res) => {
    try {
        // Kita pakai ImportirUser untuk menginisialisasi
        const { contract, gateway } = await connectToNetwork('ImportirUser');
        await contract.submitTransaction('InitWallet');
        await gateway.disconnect();
        res.json({ success: true, message: "Dompet Digital Berhasil Diinisialisasi (Importir: 10M, Petani: 0)" });
    } catch (error) { res.status(500).json({ success: false, error: error.message }); }
});

// 2. Cek Saldo
app.get('/api/wallet/:mspId', async (req, res) => {
    try {
        // mspId bisa 'ImportirMSP' atau 'PetaniMSP'
        const { contract, gateway } = await connectToNetwork('PetaniUser'); // Siapa saja bisa cek
        const result = await contract.evaluateTransaction('GetWalletBalance', req.params.mspId);
        await gateway.disconnect();
        res.json(JSON.parse(result.toString()));
    } catch (error) { res.status(500).json({ error: error.message }); }
});


// Helper -----------------------------

// GET READ (Bisa diakses siapa saja, default PetaniUser)
app.get('/api/all-batches', async (req, res) => {
    try {
        const { contract, gateway } = await connectToNetwork('PetaniUser');
        const result = await contract.evaluateTransaction('GetAllBatches');
        await gateway.disconnect();
        if (!result || result.length === 0) return res.json([]);
        res.json(JSON.parse(result.toString()));
    } catch (error) { res.status(500).json({ error: error.message }); }
});

app.get('/api/batch/:id', async (req, res) => {
    try {
        const { contract, gateway } = await connectToNetwork('PetaniUser');
        const result = await contract.evaluateTransaction('ReadBatch', req.params.id);
        await gateway.disconnect();
        if (!result || result.length === 0) return res.status(404).json({ error: "Not Found" });
        res.json(JSON.parse(result.toString()));
    } catch (error) {
        if(error.message.includes("tidak ditemukan")) return res.status(404).json({ error: "Batch Not Found" });
        res.status(500).json({ error: error.message });
    }
});

app.listen(PORT, () => {
    console.log(`Server 5-Org berjalan di http://localhost:${PORT}`);
});