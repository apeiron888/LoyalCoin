import { Lucid, Blockfrost, fromText, applyParamsToScript, Constr } from "lucid-cardano";
import fs from "fs";
import path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const BLOCKFROST_PROJECT_ID = "preprod6OurCW7t1wZmS1dHM80IOMLluKOYrOdg";
const BLOCKFROST_API_URL = "https://cardano-preprod.blockfrost.io/api/v0";
const WALLET_FILE = path.join(__dirname, "wallet.json");
const MINTING_RESULT_FILE = path.join(__dirname, "minting_result.json");

async function main() {
    console.log("🚀 LoyalCoin Minting Script (Lucid)");
    console.log("===================================");

    // 1. Initialize Lucid
    const lucid = await Lucid.new(
        new Blockfrost(BLOCKFROST_API_URL, BLOCKFROST_PROJECT_ID),
        "Preprod",
    );

    // 2. Load or Create Wallet
    let privateKey;
    if (fs.existsSync(WALLET_FILE)) {
        console.log("📂 Loading existing wallet from wallet.json");
        const walletData = JSON.parse(fs.readFileSync(WALLET_FILE, "utf8"));
        privateKey = walletData.privateKey;
    } else {
        console.log("🆕 Generating NEW wallet...");
        privateKey = lucid.utils.generatePrivateKey();
        const address = await lucid.selectWalletFromPrivateKey(privateKey).wallet.address();

        // SAVE IMMEDIATELY
        fs.writeFileSync(WALLET_FILE, JSON.stringify({ privateKey, address }, null, 2));
        console.log(`💾 Wallet saved to ${WALLET_FILE}`);
    }

    // 3. Select Wallet
    lucid.selectWalletFromPrivateKey(privateKey);
    const address = await lucid.wallet.address();
    console.log(`\n📍 Wallet Address: ${address}`);

    // 4. Check Funds
    console.log("\n💰 Checking balance...");
    const utxos = await lucid.wallet.getUtxos();
    const balance = utxos.reduce((acc, u) => acc + u.assets.lovelace, 0n);
    console.log(`   Balance: ${balance} Lovelace (${Number(balance) / 1_000_000} ADA)`);

    if (balance < 10_000_000n) { // Need at least 10 ADA for safety
        console.log("\n⚠️  INSUFFICIENT FUNDS");
        console.log(`   Please send at least 100 tADA to:`);
        console.log(`   ${address}`);
        console.log("\n   Run this script again after funding.");
        return;
    }

    // 5. Define Minting Policy
    const { paymentCredential } = lucid.utils.getAddressDetails(address);
    const mintingPolicy = lucid.utils.nativeScriptFromJson({
        type: "all",
        scripts: [
            { type: "sig", keyHash: paymentCredential.hash },
            { type: "before", slot: lucid.utils.unixTimeToSlot(Date.now() + 1000 * 60 * 60 * 24 * 365) } // 1 year
        ],
    });

    const policyId = lucid.utils.mintingPolicyToId(mintingPolicy);
    console.log(`\n📜 Policy ID: ${policyId}`);

    // 6. Mint Tokens
    const assetName = fromText("LCN");
    const unit = policyId + assetName;
    const amount = 10_000_000_000n; // 10M * 1000 (3 decimals)

    console.log(`\n🔨 Minting 10,000,000.000 LCN...`);
    console.log(`   Asset ID: ${unit}`);

    try {
        const tx = await lucid.newTx()
            .mintAssets({ [unit]: amount })
            .attachMintingPolicy(mintingPolicy)
            .validTo(Date.now() + 200000) // 3 minutesish
            .complete();

        const signedTx = await tx.sign().complete();
        const txHash = await signedTx.submit();

        console.log(`\n✅ Transaction Submitted!`);
        console.log(`   Tx Hash: ${txHash}`);
        console.log(`   Explorer: https://preprod.cardanoscan.io/transaction/${txHash}`);

        // Save result
        fs.writeFileSync(MINTING_RESULT_FILE, JSON.stringify({
            policyId,
            assetId: unit,
            txHash,
            mintedAt: new Date().toISOString()
        }, null, 2));
        console.log(`💾 Minting result saved to ${MINTING_RESULT_FILE}`);

    } catch (e) {
        console.error("\n❌ Minting Failed:", e);
    }
}

main();
