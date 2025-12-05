import { bech32 } from "bech32";
import { Lucid } from "lucid-cardano";

// Mock private key (64 bytes hex = 128 chars)
// This is a random valid ed25519 private key hex for testing
const mockHexKey = "58206f3b6f9e8e23f9c6c4c9d5f8a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f858206f3b6f9e8e23f9c6c4c9d5f8a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8";

function hexToBech32PrivateKey(hexKey) {
    console.log(`Input key length: ${hexKey.length}`);
    const cleanHex = hexKey.replace(/^0x/, '');
    console.log(`Clean hex length: ${cleanHex.length}`);
    // Convert hex to bytes
    const keyBytes = Buffer.from(cleanHex, 'hex');

    // Lucid/Cardano expects 32-byte seed for ed25519_sk
    // Go's ed25519.PrivateKey is 64 bytes (seed + pubkey)
    // We need to take the first 32 bytes
    const seedBytes = keyBytes.subarray(0, 32);
    console.log(`Seed bytes length: ${seedBytes.length}`);

    // Encode as Bech32 with 'ed25519_sk' prefix
    const words = bech32.toWords(seedBytes);
    return bech32.encode('ed25519_sk', words, 1000);
}

// Working key from wallet.json
const workingKey = "ed25519_sk1pag6n0fjn0znlad40kdv8nt7va5m4ccd2d79679s49y8ezhkt74qq4ta6k";

async function test() {
    try {
        console.log("Analyzing working key...");
        const decoded = bech32.decode(workingKey, 1000);
        console.log(`Prefix: ${decoded.prefix}`);
        const data = bech32.fromWords(decoded.words);
        console.log(`Bytes length: ${data.length}`);

        console.log("\nTesting conversion of mock key...");
        console.log("Testing conversion...");
        const bech32Key = hexToBech32PrivateKey(mockHexKey);
        console.log(`Converted Bech32: ${bech32Key.substring(0, 15)}...`);

        console.log("Initializing Lucid...");
        const lucid = await Lucid.new(null, "Preprod"); // Null provider just for key testing

        console.log("Selecting wallet...");
        lucid.selectWalletFromPrivateKey(bech32Key);
        console.log("Success!");
    } catch (e) {
        console.error("Error:", e);
    }
}

test();
