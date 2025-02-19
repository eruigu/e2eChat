package main

import (
	"sync"
	"encoding/json"
	"encoding/base64"

	"github.com/RadicalApp/libsignal-protocol-go/keys"
	"github.com/RadicalApp/libsignal-protocol-go/protocol"
	"github.com/RadicalApp/libsignal-protocol-go/session"
	"github.com/RadicalApp/libsignal-protocol-go/state"
	"github.com/RadicalApp/libsignal-protocol-go/util"
)

// SignalHandler manages Signal Protocol encryption and decryption
type SignalHandler struct {
	store         *state.InMemorySessionStore
	idKey         *keys.IdentityKeyPair
	reg           *state.InMemoryPreKeyStore
	signedPreKeys *state.InMemorySignedPreKeyStore
	mutex         sync.Mutex
}

// NewSignalHandler initializes a new SignalHandler
func NewSignalHandler() (*SignalHandler, error) {
	// Generate identity key pair
	idKey, err := keys.GenerateIdentityKeyPair()
	if err != nil {
		return nil, err
	}

	// Generate registration ID
	registrationID := util.RandomUint32()

	// Generate pre-key
	preKeyID := uint32(1)
	preKey, err := keys.GeneratePreKey(preKeyID)
	if err != nil {
		return nil, err
	}

	// Generate signed pre-key
	signedPreKeyID := uint32(1)
	signedPreKey, err := keys.GenerateSignedPreKey(idKey, signedPreKeyID)
	if err != nil {
		return nil, err
	}

	// Initialize in-memory stores
	store := state.NewInMemorySessionStore()
	preKeyStore := state.NewInMemoryPreKeyStore()
	signedPreKeyStore := state.NewInMemorySignedPreKeyStore()

	// Store keys
	preKeyStore.Store(preKeyID, preKey)
	signedPreKeyStore.Store(signedPreKeyID, signedPreKey)

	handler := &SignalHandler{
		store:         store,
		idKey:         idKey,
		reg:           preKeyStore,
		signedPreKeys: signedPreKeyStore,
	}

	return handler, nil
}

// EncryptMessage encrypts a message using Signal Protocol
func (sh *SignalHandler) EncryptMessage(recipientID string, plaintext []byte) (string, error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// Get session cipher
	cipher := session.NewCipher(sh.store, sh.idKey, recipientID)

	// Encrypt message
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return "", err
	}

	// Serialize ciphertext
	serialized, err := json.Marshal(ciphertext)
	if err != nil {
		return "", err
	}

	// Encode as base64
	return base64.StdEncoding.EncodeToString(serialized), nil
}

// DecryptMessage decrypts a Signal Protocol message
func (sh *SignalHandler) DecryptMessage(senderID string, encodedCiphertext string) ([]byte, error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// Decode base64 message
	data, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	// Deserialize ciphertext
	var ciphertext protocol.CiphertextMessage
	err = json.Unmarshal(data, &ciphertext)
	if err != nil {
		return nil, err
	}

	// Get session cipher
	cipher := session.NewCipher(sh.store, sh.idKey, senderID)

	// Decrypt message
	plaintext, err := cipher.Decrypt(&ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateNewPreKey generates a new pre-key
func (sh *SignalHandler) GenerateNewPreKey() (state.PreKeyRecord, error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	preKeyID := uint32(len(sh.reg.Store) + 1)
	preKey, err := keys.GeneratePreKey(preKeyID)
	if err != nil {
		return state.PreKeyRecord{}, err
	}

	sh.reg.Store(preKeyID, preKey)
	return preKey, nil
}

// GenerateNewSignedPreKey generates a new signed pre-key
func (sh *SignalHandler) GenerateNewSignedPreKey() (state.SignedPreKeyRecord, error) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	signedPreKeyID := uint32(len(sh.signedPreKeys.Store) + 1)
	signedPreKey, err := keys.GenerateSignedPreKey(sh.idKey, signedPreKeyID)
	if err != nil {
		return state.SignedPreKeyRecord{}, err
	}

	sh.signedPreKeys.Store(signedPreKeyID, signedPreKey)
	return signedPreKey, nil
}
