package eip712

import (
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
	for _, field := range fields {
		name := field.Name
		eip712TypeName := fmt.Sprintf("%sData", name)
		if field.Type.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field '%s' in provided schema is not a struct", name)
		}
		typeToName[field.Type] = name
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

type AquareumMessage struct {
	PrimaryType string         `json:"primaryType"`
	Domain      Domain         `json:"domain"`
	Message     map[string]any `json:"message"`
	Signature   string         `json:"signature"`
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

	msg := map[string]any{}
	innerMessage, err := ActionToMap(something)
	if err != nil {
		return nil, err
	}
	msg["data"] = innerMessage
	msg["signer"] = signer.Account.Address.String()
	msg["time"] = new(big.Int).SetInt64(time.Now().UnixMilli())
	typedData := apitypes.TypedData{
		Types:       signer.Types,
		PrimaryType: name,
		Domain:      AquareumDomain.TypedDataDomain(),
		Message:     msg,
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

	finalMessage := AquareumMessage{
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

// func (signer *EIP712Signer) Verify(bs []byte, ptr *any) error {
// 	var unverified AquareumMessage
// 	err := json.Unmarshal(bs, unverified)
// 	if err != nil {
// 		return err
// 	}
// 	sig, err := hexutil.Decode(unverified.Signature)
// 	if err != nil {
// 		return err
// 	}
// 	sig[64] -= 27
// 	typedData := apitypes.TypedData{
// 		Types:       signer.Types,
// 		Domain:      AquareumDomain.TypedDataDomain(),
// 		PrimaryType: unverified.PrimaryType,
// 		Message:     unverified.Message,
// 	}
// 	hash, _, err := apitypes.TypedDataAndHash(typedData)
// 	if err != nil {
// 		return err
// 	}
// 	rpk, err := crypto.SigToPub(hash, sig)
// 	if err != nil {
// 		return err
// 	}
// 	addr := crypto.PubkeyToAddress(*rpk)
// 	actionGenerator, ok := schema.Actions[unverified.PrimaryType]
// 	if !ok {
// 		return nil, fmt.Errorf("unknown action domain: %s", unverified.Domain)
// 	}
// 	action := actionGenerator()
// 	err = LoadMap(action, unverified.Message)
// 	if err != nil {
// 		return nil, err
// 	}
// 	addrString := fmt.Sprintf("%s", addr)
// 	if addrString != action.SignerAddress() {
// 		return nil, fmt.Errorf("incorrect signer for action! signer=%s action.signer=%s", addrString, action.SignerAddress())
// 	}
// 	return &SignedEvent{
// 		Domain:    schema.Domain,
// 		Signature: unverified.Signature,
// 		Address:   addr,
// 		Action:    action,
// 	}, nil
// }

// func (am *EIP712Signer) SignTypedData(typedData apitypes.TypedData) ([]byte, error) {
// 	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
// 	if err != nil {
// 		return nil, err
// 	}
// 	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
// 	if err != nil {
// 		return nil, err
// 	}
// 	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
// 	sighash := crypto.Keccak256(rawData)

// 	return am.signHash(sighash)
// }

// func NewEIP712Signer(opts *EIP712SignerOptions) (Signer, error) {
// 	// We don't use this parameter so let's use one that doesn't exist
// 	id := new(big.Int).SetInt64(int64(9999999999))
// 	am, err := eth.NewAccountManager(ethcommon.HexToAddress(opts.EthAccountAddr), opts.EthKeystorePath, id, opts.EthKeystorePassword)
// 	if err != nil {
// 		return nil, fmt.Errorf("error initalizing eth.AccountManager: %w", err)
// 	}
// 	err = am.Unlock(opts.EthKeystorePassword)
// 	if err != nil {
// 		return nil, fmt.Errorf("error unlcoking eth.AccountManager: %w", err)
// 	}
// 	return &EIP712Signer{
// 		PrimarySchema:  opts.PrimarySchema,
// 		Schemas:        opts.Schemas,
// 		AccountManager: am,
// 	}, nil
// }

// func (s *EIP712Signer) Sign(action Action) (*SignedEvent, error) {
// 	actionMap, err := ActionToMap(action)
// 	if err != nil {
// 		return nil, err
// 	}
// 	addrStr := fmt.Sprintf("%s", s.AccountManager.Account().Address)
// 	if actionMap["signer"] != addrStr {
// 		return nil, fmt.Errorf("address mismatch signing action, signer.address=%s, action.singer=%s", addrStr, actionMap["signer"])
// 	}
// 	actionMap["signer"] = addrStr
// 	typedData := apitypes.TypedData{
// 		Types:       s.PrimarySchema.Types,
// 		Domain:      s.PrimarySchema.Domain.TypedDataDomain(),
// 		PrimaryType: action.Type(),
// 		Message:     actionMap,
// 	}
// 	_, err = typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
// 	if err != nil {
// 		return nil, fmt.Errorf("error signing EIP712Domain: %w", err)
// 	}
// 	_, err = typedData.HashStruct(typedData.PrimaryType, typedData.Message)
// 	if err != nil {
// 		return nil, fmt.Errorf("error signing struct: %w", err)
// 	}

// 	b, err := s.AccountManager.SignTypedData(typedData)
// 	if err != nil {
// 		return nil, fmt.Errorf("error signing typed data: %w", err)
// 	}
// 	// golint wants string(b) but that gives /x1234 encoded output
// 	sig := fmt.Sprintf("%s", hexutil.Bytes(b)) //nolint:gosimple
// 	return &SignedEvent{
// 		Domain:    s.PrimarySchema.Domain,
// 		Signature: sig,
// 		Address:   s.AccountManager.Account().Address,
// 		Action:    action,
// 	}, nil
// }

// // given an unverified event from an untrusted source, verify its signature
// func (s *EIP712Signer) Verify(unverified UnverifiedEvent) (*SignedEvent, error) {
// 	// find the correct schema for this action
// 	var schema *Schema
// 	for _, s := range s.Schemas {
// 		eq, err := s.Domain.Equal(&unverified.Domain)
// 		if eq && err == nil {
// 			schema = s
// 			break
// 		}
// 	}
// 	if schema == nil {
// 		return nil, fmt.Errorf("unknown event domain: %s", unverified.Domain)
// 	}
// 	sig, err := hexutil.Decode(unverified.Signature)
// 	sig[64] -= 27
// 	if err != nil {
// 		return nil, err
// 	}
// 	typedData := apitypes.TypedData{
// 		Types:       schema.Types,
// 		Domain:      schema.Domain.TypedDataDomain(),
// 		PrimaryType: unverified.PrimaryType,
// 		Message:     unverified.Message,
// 	}
// 	hash, _, err := apitypes.TypedDataAndHash(typedData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	rpk, err := crypto.SigToPub(hash, sig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	addr := crypto.PubkeyToAddress(*rpk)
// 	actionGenerator, ok := schema.Actions[unverified.PrimaryType]
// 	if !ok {
// 		return nil, fmt.Errorf("unknown action domain: %s", unverified.Domain)
// 	}
// 	action := actionGenerator()
// 	err = LoadMap(action, unverified.Message)
// 	if err != nil {
// 		return nil, err
// 	}
// 	addrString := fmt.Sprintf("%s", addr)
// 	if addrString != action.SignerAddress() {
// 		return nil, fmt.Errorf("incorrect signer for action! signer=%s action.signer=%s", addrString, action.SignerAddress())
// 	}
// 	return &SignedEvent{
// 		Domain:    schema.Domain,
// 		Signature: unverified.Signature,
// 		Address:   addr,
// 		Action:    action,
// 	}, nil
// }

// var Schema = events.Schema{
// 	Types:  Types,
// 	Domain: Domain,
// 	Actions: map[string]func() events.Action{
// 		"ChannelDefinition": func() events.Action {
// 			return &ChannelDefinition{}
// 		},
// 	},
// }
