package eip712test

import (
	"context"
	"os"
	"path/filepath"

	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"
)

// package for setting up a test wallet

var wallet = `{
  "address": "295481766f43bb048aec5d71f3bf76fdacea78f2",
  "crypto": {
    "cipher": "aes-128-ctr",
    "ciphertext": "2cd9bfb58a6d7720d064dffdbd9840f055f7e877396364ce6fdb15d496166cb6",
    "cipherparams": { "iv": "f6f53d78aac9a1af96fbbc217d17394b" },
    "kdf": "scrypt",
    "kdfparams": {
      "dklen": 32,
      "n": 262144,
      "p": 1,
      "r": 8,
      "salt": "3f66908f20dd26f98b0347f6ab6cc2e5658b751b89dcb487be59e1ba1d0b76e5"
    },
    "mac": "a284cc5c66d38cf058ee3aed52012d7375b3f463abde58a174251a36d09ea8e2"
  },
  "id": "86ec124c-ebe6-4100-811c-22396e10abe8",
  "version": 3
}`

// creates a test wallet, cleaned up after the function ends
func WithTestSigner(fn func(*eip712.EIP712Signer)) {
	dname, err := os.MkdirTemp("", "sampledir")
	defer os.RemoveAll(dname)
	fname := filepath.Join(dname, "wallet.json")
	os.WriteFile(fname, []byte(wallet), 0600)
	if err != nil {
		panic(err)
	}
	schema, err := v0.MakeV0Schema()
	if err != nil {
		panic(err)
	}
	signer, err := eip712.MakeEIP712Signer(context.Background(), &eip712.EIP712SignerOptions{
		EthKeystorePassword: "aquareumaquareum",
		EthKeystorePath:     dname,
		EthAccountAddr:      "0x295481766F43bb048Aec5D71f3Bf76FDaCEA78f2",
		Schema:              schema,
	})
	if err != nil {
		panic(err)
	}
	fn(signer)
}
