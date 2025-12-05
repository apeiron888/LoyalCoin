import { Lucid, Blockfrost, fromText, C, fromHex } from "lucid-cardano";
import fs from "fs";
import path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const BLOCKFROST_PROJECT_ID = "preprod6OurCW7t1wZmS1dHM80IOMLluKOYrOdg";
const BLOCKFROST_API_URL = "https://cardano-preprod.blockfrost.io/api/v0";

async function main() {
    let input = "";
    try {
        // Read input from stdin
        input = fs.readFileSync(0, "utf-8");
        const { privateKey, toAddress, amount, assetId } = JSON.parse(input);

        if (!privateKey || !toAddress || !amount) {
            throw new Error("Missing required fields: privateKey, toAddress, amount");
        }

        // Initialize Lucid
        const lucid = await Lucid.new(
            new Blockfrost(BLOCKFROST_API_URL, BLOCKFROST_PROJECT_ID),
            "Preprod",
        );

        // Handle Go-generated Ed25519 keys (64 bytes / 128 hex chars)
        let effectivePrivateKey = privateKey;
        if (!privateKey.startsWith("ed25519_sk") && privateKey.length === 128) {
            effectivePrivateKey = privateKey.slice(0, 64);
        }

        // Select wallet using CML to convert to Bech32
        try {
            const priv = C.PrivateKey.from_normal_bytes(fromHex(effectivePrivateKey));
            const bech32Key = priv.to_bech32();
            lucid.selectWalletFromPrivateKey(bech32Key);
        } catch (e) {
            console.error(JSON.stringify({ status: "error", message: "Failed to initialize wallet from private key", stack: e.stack }));
            process.exit(1);
        }

        // Build transaction
        let tx = lucid.newTx();

        if (assetId && assetId !== "lovelace") {
            // Transfer native asset
            // Amount is in atomic units (BigInt)
            const assets = { [assetId]: BigInt(amount) };
            tx = tx.payToAddress(toAddress, assets);
        } else {
            // Transfer ADA
            tx = tx.payToAddress(toAddress, { lovelace: BigInt(amount) });
        }

        // Complete, sign, and submit
        const completedTx = await tx.complete();
        const signedTx = await completedTx.sign().complete();
        const txHash = await signedTx.submit();

        // Output result
        console.log(JSON.stringify({ status: "ok", txHash: txHash }));

    } catch (error) {
        const msg = error.message || String(error);
        const stack = error.stack || "No stack trace";
        console.error(JSON.stringify({ status: "error", message: msg, stack: stack }));
        process.exit(1);
    }
}

main();
