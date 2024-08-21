package main

import (
	"github.com/ThalesIgnite/crypto11"
)

func Run() error {
	sc, err := crypto11.Configure(&crypto11.Config{
		TokenLabel: "C2PA Signer",
		Path:       "/opt/homebrew/Cellar/opensc/0.25.1/lib/opensc-pkcs11.so",
		Pin:        "99999999",
	})
	if err != nil {
		return err
	}

	// bs := make([]byte, 4)
	// binary.LittleEndian.PutUint32(bs, 2)
	// signer, err := sc.FindKeyPair(nil, bs)
	// if err != nil {
	// 	return err
	// }
	// if signer == nil {
	// 	return fmt.Errorf("keypair not found")
	// }
	// pub := signer.Public()
	// fmt.Println("%v", pub)
	signers, err := sc.FindAllKeyPairs()
	if err != nil {
		return err
	}
	signer := signers[0]
	// for _, signer := range signers {
	// 	pub := signer.Public()
	// 	fmt.Println("%v", pub)
	// }
	return nil
}

func main() {
	err := Run()
	if err != nil {
		panic(err)
	}
}
