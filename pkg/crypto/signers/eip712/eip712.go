package eip712

import (
	"bytes"
	"context"
	gocrypto "crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"

	"aquareum.tv/aquareum/pkg/log"
	_ "aquareum.tv/aquareum/pkg/media/mediatesting"
	"aquareum.tv/aquareum/pkg/schema"
	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// type Signer interface {
// 	Sign(action Action) (*SignedEvent, error)
// 	Verify(unverified UnverifiedEvent) (*SignedEvent, error)
// }

// schema-aware signer for signing actions and verifying untrusted payloads
// type Signer interface {
// 	Sign(action Action) (*SignedEvent, error)
// 	Verify(unverified UnverifiedEvent) (*SignedEvent, error)
// }

// Signer implemented with EIP712
type EIP712Signer struct {
	// // When I sign an action, which schema should I use?
	// PrimarySchema *Schema
	// // All supported schemas for verification purposes
	// Schemas []*Schema
	// // Eth Account Manager
	// AccountManager eth.AccountManager
	KeyStore     *keystore.KeyStore
	Account      *accounts.Account
	Opts         *EIP712SignerOptions
	EIP712Schema *schema.EIP712SchemaStruct
}

type EIP712SignerOptions struct {
	// PrimarySchema       *Schema
	// Schemas             []*Schema
	EthKeystorePassword string
	EthKeystorePath     string
	EthAccountAddr      string
	Schema              schema.Schema
}

func MakeEIP712Signer(ctx context.Context, opts *EIP712SignerOptions) (*EIP712Signer, error) {

	eip712Schema, err := opts.Schema.EIP712()
	if err != nil {
		return nil, err
	}
	signer := &EIP712Signer{
		Opts:         opts,
		EIP712Schema: eip712Schema,
	}

	if opts.EthKeystorePath != "" {
		err := signer.InitKeystore(ctx)
		if err != nil {
			return nil, err
		}
		log.Log(ctx, "successfully initalized keystore", "keystorePath", opts.EthKeystorePath, "address", signer.Hex())
	} else {
		log.Log(ctx, "my EthKeystorePath is empty; EIP-712 signing won't work (which is fine, i guess)")
	}

	return signer, nil
}

func (signer *EIP712Signer) InitKeystore(ctx context.Context) error {
	keyStore := keystore.NewKeyStore(signer.Opts.EthKeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	var account *accounts.Account
	if signer.Opts.EthAccountAddr != "" {
		addr := common.HexToAddress(signer.Opts.EthAccountAddr)

		acctExists := keyStore.HasAddress(addr)
		if !acctExists {
			return fmt.Errorf("keystore does not contain account %s", signer.Opts.EthAccountAddr)
		}
		for _, a := range keyStore.Accounts() {
			if a.Address == addr {
				account = &a
			}
		}
		if account == nil {
			return fmt.Errorf("keystore does not contain account %s", signer.Opts.EthAccountAddr)
		}
	} else {
		count := len(keyStore.Accounts())
		if count > 1 {
			return fmt.Errorf("keystore contains more than one account; specify which one to use with -eth-account-addr")
		}
		if count == 1 {
			account = &keyStore.Accounts()[0]
		}
		if count == 0 {
			acct, err := keyStore.NewAccount(signer.Opts.EthKeystorePassword)
			if err != nil {
				return fmt.Errorf("unable to generate new ethereum account: %w", err)
			}
			account = &acct
			log.Log(ctx, "generated new ethereum key", "addr", signer.Hex())
		}
	}
	err := keyStore.Unlock(*account, signer.Opts.EthKeystorePassword)
	if err != nil {
		return err
	}
	signer.Account = account
	signer.KeyStore = keyStore
	return nil
}

// return account address as a hex string
func (signer *EIP712Signer) Hex() string {
	return hexutil.Encode(signer.Account.Address.Bytes())
}

func (signer *EIP712Signer) KnownTypes() []string {
	types := []string{}
	for _, val := range signer.EIP712Schema.TypeToName {
		types = append(types, val)
	}
	return types
}

type SignedMessage interface {
	Signer() string
	Time() int64
	Data() any
}
type AquareumEIP712 struct {
	PrimaryType string                    `json:"primaryType"`
	Domain      *apitypes.TypedDataDomain `json:"domain"`
	Message     AquareumEIP712Message     `json:"message"`
	Signature   string                    `json:"signature"`
}

type AquareumEIP712Message struct {
	MsgSigner string `json:"signer"`
	MsgTime   int64  `json:"time"`
	MsgData   any    `json:"data"`
}

// return a Map representation suitable for passing to the geth functions
func (msg AquareumEIP712Message) Map() map[string]any {
	m := map[string]any{}
	m["signer"] = msg.MsgSigner
	m["time"] = new(big.Int).SetInt64(msg.MsgTime)
	m["data"] = msg.MsgData
	return m
}

func (msg *AquareumEIP712Message) Signer() string {
	return msg.MsgSigner
}

func (msg *AquareumEIP712Message) Time() int64 {
	return msg.MsgTime
}

func (msg *AquareumEIP712Message) Data() any {
	return msg.MsgData
}

func (signer *EIP712Signer) SignMessage(something any) ([]byte, error) {
	typ := reflect.TypeOf(something)
	name, ok := signer.EIP712Schema.TypeToName[typ]
	if !ok {
		allTypes := strings.Join(signer.KnownTypes(), ", ")
		return nil, fmt.Errorf("unknown type provided to Sign, expected one of [%s]", allTypes)
	}

	innerMessage, err := ActionToMap(something)
	if err != nil {
		return nil, err
	}
	msg := AquareumEIP712Message{
		MsgData:   innerMessage,
		MsgSigner: signer.Account.Address.String(),
		MsgTime:   time.Now().UnixMilli(),
	}
	typedData := apitypes.TypedData{
		Types:       signer.EIP712Schema.Types,
		PrimaryType: name,
		Domain:      *signer.EIP712Schema.Domain,
		Message:     msg.Map(),
	}
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	rawHash := crypto.Keccak256(rawData)
	sig, err := signer.KeyStore.SignHash(*signer.Account, rawHash)
	if err != nil {
		return nil, fmt.Errorf("error calling KeyStore.SignHash: %w", err)
	}

	// sig is in the [R || S || V] format where V is 0 or 1
	// Convert the V param to 27 or 28
	v := sig[64]
	if v == byte(0) || v == byte(1) {
		v += 27
	}
	sig = append(sig[:64], v)
	// golint wants string(b) but that gives /x1234 encoded output
	sigHex := hexutil.Bytes(sig).String()

	finalMessage := AquareumEIP712{
		PrimaryType: name,
		Domain:      signer.EIP712Schema.Domain,
		Message:     msg,
		Signature:   sigHex,
	}

	data, err := json.Marshal(finalMessage)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ActionToMap(a any) (map[string]any, error) {
	data, err := json.Marshal(a)

	if err != nil {
		return nil, err
	}

	newMap := map[string]any{}
	err = json.Unmarshal(data, &newMap)
	if err != nil {
		return nil, err
	}
	return newMap, nil
}

func (signer *EIP712Signer) Verify(bs []byte) (SignedMessage, error) {
	var unverified AquareumEIP712
	err := json.Unmarshal(bs, &unverified)
	if err != nil {
		return nil, fmt.Errorf("error on json.Unmarshal: %w", err)
	}
	sig, err := hexutil.Decode(unverified.Signature)
	if err != nil {
		return nil, fmt.Errorf("error on hexutil.Decode: %w", err)
	}
	sig[64] -= 27
	typedData := apitypes.TypedData{
		Types:       signer.EIP712Schema.Types,
		Domain:      *signer.EIP712Schema.Domain,
		PrimaryType: unverified.PrimaryType,
		Message:     unverified.Message.Map(),
	}
	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("error on apitypes.TypedDataAndHash: %w", err)
	}
	rpk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return nil, fmt.Errorf("error on crypto.SigToPub: %w", err)
	}
	addr := crypto.PubkeyToAddress(*rpk)
	messageSignerAddr, err := hexutil.Decode(unverified.Message.Signer())
	if err != nil {
		return nil, fmt.Errorf("error on hexutil.Decode: %w", err)
	}
	if !bytes.Equal(messageSignerAddr, addr.Bytes()) {
		specifiedSigner := hexutil.Encode(messageSignerAddr)
		actualSigner := hexutil.Encode(addr.Bytes())
		return nil, fmt.Errorf("message signature does not match signer on message specified=%s actual=%s", specifiedSigner, actualSigner)
	}
	typ, ok := signer.EIP712Schema.NameToType[unverified.PrimaryType]
	if !ok {
		return nil, fmt.Errorf("go type not found for message type %s", unverified.PrimaryType)
	}
	dataBs, err := json.Marshal(unverified.Message.Data())
	if err != nil {
		return nil, err
	}
	something := reflect.New(typ).Interface()
	err = json.Unmarshal(dataBs, something)
	if err != nil {
		return nil, err
	}
	// new object that has the correct type hidden within!
	signed := AquareumEIP712Message{
		MsgSigner: unverified.Message.Signer(),
		MsgTime:   unverified.Message.Time(),
		MsgData:   something,
	}
	return &signed, nil
}

func (signer *EIP712Signer) Sign(rand io.Reader, digest []byte, opts gocrypto.SignerOpts) (signature []byte, err error) {
	sig, err := signer.EthSign(digest)

	if err != nil {
		return nil, err
	}

	// strip off the ethereum-style recovery bit for use elesewhere in the world
	return sig[:64], nil
}

// sign with an ethereum-style recovery bit
func (signer *EIP712Signer) EthSign(digest []byte) (signature []byte, err error) {
	sig, err := signer.KeyStore.SignHash(*signer.Account, digest)
	if err != nil {
		return []byte{}, err
	}
	v := sig[64]
	if v == byte(0) || v == byte(1) {
		v += 27
	}
	sig = append(sig[:64], v)
	return sig, nil
}

func (signer *EIP712Signer) Public() gocrypto.PublicKey {
	key, err := signer.public()
	if err != nil {
		panic(err)
	}
	return key
}

// public helper that returns an error instead of panic
func (signer *EIP712Signer) public() (gocrypto.PublicKey, error) {
	nullhash := make([]byte, 32)
	sig, err := signer.EthSign(nullhash)
	if err != nil {
		return nil, fmt.Errorf("error getting public key from signer.Sign: %w", err)
	}
	sig[64] -= 27
	rpk, err := crypto.SigToPub(nullhash, sig)
	if err != nil {
		return nil, fmt.Errorf("error getting public key from crypto.SigToPub: %w", err)
	}
	return rpk, nil
}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

type signerThatLies struct {
	trueSigner *EIP712Signer
	falseKey   *ecdsa.PrivateKey
}

func (signer *signerThatLies) Public() gocrypto.PublicKey {
	return signer.falseKey.Public()
}

func (signer *signerThatLies) Sign(rand io.Reader, digest []byte, opts gocrypto.SignerOpts) (signature []byte, err error) {
	return signer.trueSigner.Sign(rand, digest, opts)
}

func (signer *EIP712Signer) GenerateCert() ([]byte, error) {
	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature

	notBefore := time.Now()
	notAfter := notBefore.Add((100 * 365 * 24) * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate P256 key: %w", err)
	}

	// pub := priv.Public().(*ecdsa.PublicKey)
	// publicKeyBytes := elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
	pub := signer.Public().(*ecdsa.PublicKey)
	publicKeyBytes := elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
	idhash := sha1.Sum(publicKeyBytes)
	subjectKeyId := idhash[:]

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: signer.Hex(),
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection},
		BasicConstraintsValid: true,
		SubjectKeyId:          subjectKeyId,
		AuthorityKeyId:        subjectKeyId,
	}

	p256DERBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create p256 certificate: %w", err)
	}

	var p256cert certificate
	if _, err := asn1.Unmarshal(p256DERBytes, &p256cert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal p256 cert: %w", err)
	}

	old := p256cert.TBSCertificate

	var paramBytes []byte
	paramBytes, err = asn1.Marshal(c2pa.OID_SECP256K1)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OID_SECP256K1: %w", err)

	}
	var signingAlg pkix.AlgorithmIdentifier
	signingAlg.Algorithm = old.PublicKey.Algorithm.Algorithm
	signingAlg.Parameters.FullBytes = paramBytes

	k256info := publicKeyInfo{
		Algorithm: signingAlg,
		PublicKey: asn1.BitString{
			Bytes:     publicKeyBytes,
			BitLength: len(publicKeyBytes) * 8,
		},
	}

	tbs := tbsCertificate{
		Version:            old.Version,
		SerialNumber:       old.SerialNumber,
		SignatureAlgorithm: old.SignatureAlgorithm,
		Issuer:             old.Issuer,
		Validity:           old.Validity,
		Subject:            old.Subject,
		PublicKey:          k256info,
		UniqueId:           old.UniqueId,
		SubjectUniqueId:    old.SubjectUniqueId,
		Extensions:         old.Extensions,
	}

	toSign, err := asn1.Marshal(tbs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal p256 cert: %w", err)
	}
	h := gocrypto.SHA256.New()
	h.Write(toSign)
	digest := h.Sum(nil)
	sig, err := signer.Sign(rand.Reader, digest, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign k256k cert: %w", err)
	}
	k256cert := certificate{
		TBSCertificate:     tbs,
		SignatureAlgorithm: p256cert.SignatureAlgorithm,
		SignatureValue:     asn1.BitString{Bytes: sig, BitLength: len(sig) * 8},
	}

	k256DERBytes, err := asn1.Marshal(k256cert)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal k256k cert: %w", err)
	}

	bs := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: k256DERBytes})
	return bs, nil
}

type certificate struct {
	TBSCertificate     tbsCertificate
	SignatureAlgorithm pkix.AlgorithmIdentifier
	SignatureValue     asn1.BitString
}

type tbsCertificate struct {
	Raw                asn1.RawContent
	Version            int `asn1:"optional,explicit,default:0,tag:0"`
	SerialNumber       *big.Int
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Issuer             asn1.RawValue
	Validity           validity
	Subject            asn1.RawValue
	PublicKey          publicKeyInfo
	UniqueId           asn1.BitString   `asn1:"optional,tag:1"`
	SubjectUniqueId    asn1.BitString   `asn1:"optional,tag:2"`
	Extensions         []pkix.Extension `asn1:"omitempty,optional,explicit,tag:3"`
}

type dsaAlgorithmParameters struct {
	P, Q, G *big.Int
}

type validity struct {
	NotBefore, NotAfter time.Time
}

type publicKeyInfo struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}
