package eip712

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

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
	Types      apitypes.Types
	TypeToName map[reflect.Type]string
	NameToType map[string]reflect.Type
	KeyStore   *keystore.KeyStore
	Account    *accounts.Account
}

type EIP712SignerOptions struct {
	// PrimarySchema       *Schema
	// Schemas             []*Schema
	EthKeystorePassword string
	EthKeystorePath     string
	EthAccountAddr      string
	Schema              any
}

var AquareumDomain = Domain{
	Name:    "Aquareum",
	Version: "0.0.1",
}

func MakeEIP712Signer(opts *EIP712SignerOptions) (*EIP712Signer, error) {
	keyStore := keystore.NewKeyStore(opts.EthKeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	addr := common.HexToAddress(opts.EthAccountAddr)

	acctExists := keyStore.HasAddress(addr)
	if !acctExists {
		return nil, fmt.Errorf("keystore does not contain account %s", opts.EthAccountAddr)
	}
	var account *accounts.Account
	for _, a := range keyStore.Accounts() {
		if a.Address == addr {
			account = &a
		}
	}
	if account == nil {
		return nil, fmt.Errorf("keystore does not contain account %s", opts.EthAccountAddr)
	}
	err := keyStore.Unlock(*account, opts.EthKeystorePassword)
	if err != nil {
		return nil, err
	}

	var eip712Types = apitypes.Types{
		"EIP712Domain": {
			{
				Name: "name",
				Type: "string",
			},
			{
				Name: "version",
				Type: "string",
			},
		},
	}

	stype := reflect.TypeOf(opts.Schema)
	if stype.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema parameter of MakeEIP712Signer is not a struct")
	}
	fields := reflect.VisibleFields(stype)
	typeToName := map[reflect.Type]string{}
	nameToType := map[string]reflect.Type{}
	for _, field := range fields {
		name := field.Name
		eip712TypeName := fmt.Sprintf("%sData", name)
		if field.Type.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field '%s' in provided schema is not a struct", name)
		}
		typeToName[field.Type] = name
		nameToType[name] = field.Type
		parentType := []apitypes.Type{
			{
				Name: "signer",
				Type: "address",
			},
			{
				Name: "time",
				Type: "int64",
			},
			{
				Name: "data",
				Type: eip712TypeName,
			},
		}
		typeSlice := []apitypes.Type{}

		subfields := reflect.VisibleFields(field.Type)
		for _, subfield := range subfields {
			eipType, err := goToEIP712(subfield)
			if err != nil {
				return nil, fmt.Errorf("error handling type %s: %w", name, err)
			}
			typeSlice = append(typeSlice, eipType)
		}
		eip712Types[name] = parentType
		eip712Types[eip712TypeName] = typeSlice
	}

	return &EIP712Signer{
		Types:      eip712Types,
		KeyStore:   keyStore,
		Account:    account,
		TypeToName: typeToName,
		NameToType: nameToType,
	}, nil
}

func (signer *EIP712Signer) KnownTypes() []string {
	types := []string{}
	for _, val := range signer.TypeToName {
		types = append(types, val)
	}
	return types
}

// turns a go type into an eip712 type
func goToEIP712(field reflect.StructField) (apitypes.Type, error) {
	var typ string
	kind := field.Type.Kind()
	if kind == reflect.String {
		typ = "string"
	} else if kind == reflect.Int64 {
		typ = "int64"
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return apitypes.Type{}, fmt.Errorf("could not find field name for %s", field.Name)
	}
	return apitypes.Type{
		Name: jsonTag,
		Type: typ,
	}, nil
}

type Domain struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SignedMessage interface {
	Signer() string
	Time() int64
	Data() any
}
type AquareumEIP712 struct {
	PrimaryType string                `json:"primaryType"`
	Domain      Domain                `json:"domain"`
	Message     AquareumEIP712Message `json:"message"`
	Signature   string                `json:"signature"`
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

// convert to a TypedDataDomain suitable for signing by eth tooling
func (d *Domain) TypedDataDomain() apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:    d.Name,
		Version: d.Version,
	}
}

func (signer *EIP712Signer) Sign(something any) ([]byte, error) {
	typ := reflect.TypeOf(something)
	name, ok := signer.TypeToName[typ]
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
		Types:       signer.Types,
		PrimaryType: name,
		Domain:      AquareumDomain.TypedDataDomain(),
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
		Domain:      AquareumDomain,
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
		Types:       signer.Types,
		Domain:      AquareumDomain.TypedDataDomain(),
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
		return nil, fmt.Errorf("message signature does not match signer on message")
	}
	typ, ok := signer.NameToType[unverified.PrimaryType]
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
