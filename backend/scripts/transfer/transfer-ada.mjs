import { Lucid, Blockfrost } from "lucid-cardano";
import { bech32 } from "bech32";

const BLOCKFROST_PROJECT_ID = process.env.BLOCKFROST_PROJECT_ID || "preprod6OurCW7t1wZmS1dHM80IOMLluKOYrOdg";
const BLOCKFROST_API_URL = process.env.BLOCKFROST_API_URL || "https://cardano-preprod.blockfrost.io/api/v0";

// Convert hex private key to Bech32 (ed25519_sk1...) format
function hexToBech32PrivateKey(hexKey) {
    // Remove any '0x' prefix if present
    const cleanHex = hexKey.replace(/^0x/, '');

    // Convert hex to bytes
    const keyBytes = Buffer.from(cleanHex, 'hex');

    // Lucid/Cardano expects 32-byte seed for ed25519_sk
    // Go's ed25519.PrivateKey is 64 bytes (seed + pubkey)
    // We need to take the first 32 bytes
    const seedBytes = keyBytes.subarray(0, 32);

    // Encode as Bech32 with 'ed25519_sk' prefix
    // Convert 8-bit data to 5-bit groups
    const words = bech32.toWords(seedBytes);

    // Encode with ed25519_sk prefix
    return bech32.encode('ed25519_sk', words, 1000); // 1000 is max length
}


// Helper to sleep
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Helper to retry with exponential backoff
async function retryWithBackoff(fn, maxRetries = 3, baseDelay = 1000) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await fn();
        } catch (error) {
            if (attempt === maxRetries) throw error;
            const delay = baseDelay * Math.pow(2, attempt - 1);
            await sleep(delay);
        }
    }
}

async function main() {
    // Read input from stdin
    const chunks = [];
    for await (const chunk of process.stdin) chunks.push(chunk);
    const input = JSON.parse(Buffer.concat(chunks).toString());

    const { privateKey, toAddress, lovelace } = input;

    try {
        // Initialize Lucid with retry logic (Blockfrost can be slow)
        const lucid = await retryWithBackoff(async () => {
            return await Lucid.new(
                new Blockfrost(BLOCKFROST_API_URL, BLOCKFROST_PROJECT_ID),
                "Preprod",
            );
        }, 3, 2000); // 3 retries, starting with 2s delay

        // Convert hex private key to Bech32 format (ed25519_sk1...)
        // Lucid requires keys in this specific format
        const bech32Key = hexToBech32PrivateKey(privateKey);

        // Select wallet with Bech32-encoded private key
        lucid.selectWalletFromPrivateKey(bech32Key);

        // Sync wallet to get fresh UTXOs from Blockfrost
        // This ensures we don't use stale UTXO data from a previous transaction
        const utxos = await retryWithBackoff(async () => {
            return await lucid.wallet.getUtxos();
        }, 3, 1000);

        if (!utxos || utxos.length === 0) {
            throw new Error("No UTXOs available in wallet. The wallet may be unfunded or previous transactions are still pending.");
        }

        // Log available UTXOs for debugging
        const totalAvailable = utxos.reduce((sum, u) => sum + u.assets.lovelace, 0n);

        // Verify we have enough funds
        const requiredLovelace = BigInt(lovelace) + 500000n; // Amount + estimated fee
        if (totalAvailable < requiredLovelace) {
            throw new Error(`Insufficient funds. Required: ${requiredLovelace}, Available: ${totalAvailable}`);
        }

        // Build and submit simple ADA transfer
        const tx = await retryWithBackoff(async () => {
            return await lucid.newTx()
                .payToAddress(toAddress, { lovelace: BigInt(lovelace) })
                .complete();
        }, 2, 2000);

        const signedTx = await tx.sign().complete();
        const txHash = await signedTx.submit();

        console.log(JSON.stringify({
            status: "ok",
            txHash: txHash,
        }));
    } catch (error) {
        // Write error to stderr (not stdout) so backend can capture it
        console.error(JSON.stringify({
            status: "error",
            message: error.message || String(error),
            stack: error.stack || "No stack trace"
        }));
        process.exit(1);
    }
}

main();
