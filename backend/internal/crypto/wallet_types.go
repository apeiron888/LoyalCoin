package crypto

type WalletResult struct {
	Address          string
	EncryptedPrivKey string
	WrappedDEK       string
	PubKeyHex        string
}
