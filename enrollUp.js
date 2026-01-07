/* enrollUp.js - 5 ORGANISASI */
const FabricCAServices = require('fabric-ca-client');
const { Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

async function setupOrg(orgName, mspId, userId, wallet) {
    try {
        const ccpPath = path.resolve(__dirname, `fablo-target/fabric-config/connection-profiles/connection-profile-${orgName.toLowerCase()}.json`);
        
        if (!fs.existsSync(ccpPath)) {
            console.error(`❌ CCP untuk ${orgName} tidak ditemukan di: ${ccpPath}`);
            return;
        }

        const ccpContent = fs.readFileSync(ccpPath, 'utf8');
        const ccp = JSON.parse(ccpContent.replace(/http:\/\//g, 'https://'));

        const caKey = Object.keys(ccp.certificateAuthorities)[0];
        const caInfo = ccp.certificateAuthorities[caKey];
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

        console.log(`\n--- Processing Organization: ${orgName} (${mspId}) ---`);

        // 1. Enroll Admin
        const adminName = `admin-${orgName.toLowerCase()}`;
        let identity = await wallet.get(adminName);
        
        if (identity) {
            console.log(`   ℹ️  Admin ${adminName} sudah ada.`);
        } else {
            const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
            const x509Identity = {
                credentials: { certificate: enrollment.certificate, privateKey: enrollment.key.toBytes() },
                mspId: mspId, type: 'X.509',
            };
            await wallet.put(adminName, x509Identity);
            console.log(`   ✅ Sukses Enroll Admin: ${adminName}`);
        }

        // 2. Register & Enroll User Aplikasi
        const userIdentity = await wallet.get(userId);
        if (userIdentity) {
            console.log(`   ℹ️  User ${userId} sudah ada di wallet.`);
        } else {
            try {
                // Load Admin untuk register user
                const adminUserIdentity = await wallet.get(adminName);
                const provider = wallet.getProviderRegistry().getProvider(adminUserIdentity.type);
                const adminUserContext = await provider.getUserContext(adminUserIdentity, 'admin');

                // Register (Langkah ini gagal jika user sudah ada di CA)
                const secret = await ca.register({ enrollmentID: userId, role: 'client' }, adminUserContext);
                
                // Jika register sukses, langsung Enroll
                const enrollment = await ca.enroll({ enrollmentID: userId, enrollmentSecret: secret });
                const x509UserIdentity = {
                    credentials: { certificate: enrollment.certificate, privateKey: enrollment.key.toBytes() },
                    mspId: mspId, type: 'X.509',
                };
                await wallet.put(userId, x509UserIdentity);
                console.log(`   ✅ Sukses Enroll User (Registered): ${userId}`);

            } catch (registerError) {
                // --- LOGIKA PERBAIKAN ERROR 74 (Identity already registered) ---
                if (registerError.message.includes('code: 74')) {
                    console.log(`   ⚠️ User ${userId} sudah terdaftar di CA. Mencoba Enroll ulang...`);
                    
                    // Karena ini development, kita asumsikan secret = user ID untuk enrollment
                    const enrollment = await ca.enroll({ enrollmentID: userId, enrollmentSecret: userId }); 

                    if(enrollment.certificate) {
                        const x509UserIdentity = {
                            credentials: { certificate: enrollment.certificate, privateKey: enrollment.key.toBytes() },
                            mspId: mspId, type: 'X.509',
                        };
                        await wallet.put(userId, x509UserIdentity);
                        console.log(`   ✅ Sukses Enroll User (Re-enrolled): ${userId}`);
                    } else {
                         console.log(`   ❌ Gagal Enroll Ulang User: ${userId}. Sertifikat tidak ditemukan.`);
                    }

                } else {
                    throw registerError;
                }
            }
        }

    } catch (error) {
        console.error(`❌ Gagal Setup ${orgName}: ${error.message}`);
    }
}

async function main() {
    try {
        const walletPath = path.join(process.cwd(), 'wallet');
        
        // Hapus wallet lama agar bersih
        if (fs.existsSync(walletPath)) {
            fs.rmSync(walletPath, { recursive: true, force: true });
            console.log("  Wallet lama dihapus.");
        }
        
        const wallet = await Wallets.newFileSystemWallet(walletPath);

        // --- SETUP 5 ORGANISASI ---
        await setupOrg('Petani',    'PetaniMSP',    'PetaniUser',    wallet);
        await setupOrg('Logistik',  'LogistikMSP',  'LogistikUser',  wallet);
        await setupOrg('Koperasi',  'KoperasiMSP',  'KoperasiUser',  wallet); 
        await setupOrg('Regulator', 'RegulatorMSP', 'RegulatorUser', wallet);
        await setupOrg('Importir',  'ImportirMSP',  'ImportirUser',  wallet);

        console.log('\n SELESAI. Semua 5 Organisasi Siap!');

    } catch (error) {
        console.error(`Error Fatal: ${error}`);
    }
}

main();