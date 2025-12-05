import { Lucid, Blockfrost } from "lucid-cardano";
import fs from "fs";
import path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const BLOCKFROST_PROJECT_ID = "preprod6OurCW7t1wZmS1dHM80IOMLluKOYrOdg";
const BLOCKFROST_API_URL = "https://cardano-preprod.blockfrost.io/api/v0";
const WALLET_FILE = path.join(__dirname, "wallet.json");
const MINTING_RESULT_FILE = path.join(__dirname, "minting_result.json");

// Admin wallet address - CHANGE THIS to your admin wallet
const ADMIN_ADDRESS = "addr_test1vzewj0wvwda4z5w3ezmf7sa6fu3xzn2nqkhvylyxvd0u6sgvz7xtp";
const AMOUNT_TO_SEND = 10_000_000_000n; // 10M LCN (10,000,000.000 with 3 decimals)

async function main() {
    console.log("📤 Sending LCN to Admin Wallet");
    console.log("================================");

    // Load minting result to get asset ID
    const mintingResult = JSON.parse(fs.readFileSync(MINTING_RESULT_FILE, "utf8"));
    const assetId = mintingResult.assetId;
    console.log(`   Asset ID: ${assetId}`);
    console.log(`   Admin Address: ${ADMIN_ADDRESS}`);

    // Initialize Lucid
    const lucid = await Lucid.new(
        new Blockfrost(BLOCKFROST_API_URL, BLOCKFROST_PROJECT_ID),
        "Preprod",
    );

    // Load wallet
    const walletData = JSON.parse(fs.readFileSync(WALLET_FILE, "utf8"));
    lucid.selectWalletFromPrivateKey(walletData.privateKey);

    // Build transaction
    console.log(`\n💸 Sending ${Number(AMOUNT_TO_SEND) / 1000} LCN to admin...`);

    const tx = await lucid.newTx()
        .payToAddress(ADMIN_ADDRESS, { [assetId]: AMOUNT_TO_SEND, lovelace: 2_000_000n }) // Include 2 ADA for min UTXO
        .complete();

    const signedTx = await tx.sign().complete();
    const txHash = await signedTx.submit();

    console.log(`\n✅ Transfer Submitted!`);
    console.log(`   Tx Hash: ${txHash}`);
    console.log(`   Explorer: https://preprod.cardanoscan.io/transaction/${txHash}`);
    console.log(`\n⏳ Wait a few moments, then refresh the Admin Dashboard to see the balance update.`);
}

main().catch(console.error);
