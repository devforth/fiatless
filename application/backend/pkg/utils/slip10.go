package utils

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/crypto/ed25519"
)

const slip10SeedKey = "ed25519 seed"

// HardenedIndex returns i | 0x80000000
func HardenedIndex(i uint32) uint32 { return i | 0x80000000 }

type Slip10Node struct {
	Key       []byte // 32-byte private key seed
	ChainCode []byte // 32-byte chain code
}

func NewSlip10Master(seed []byte) (*Slip10Node, error) {
	mac := hmac.New(sha512.New, []byte(slip10SeedKey))
	if _, err := mac.Write(seed); err != nil {
		return nil, err
	}
	i := mac.Sum(nil)
	return &Slip10Node{Key: i[:32], ChainCode: i[32:]}, nil
}

func (n *Slip10Node) DeriveHardened(i uint32) (*Slip10Node, error) {
	if i&0x80000000 == 0 {
		return nil, errors.New("slip10 ed25519 supports only hardened indices")
	}
	data := make([]byte, 0, 1+32+4)
	data = append(data, 0x00)
	data = append(data, n.Key...)
	var be [4]byte
	binary.BigEndian.PutUint32(be[:], i)
	data = append(data, be[:]...)
	mac := hmac.New(sha512.New, n.ChainCode)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	o := mac.Sum(nil)
	return &Slip10Node{Key: o[:32], ChainCode: o[32:]}, nil
}

func (n *Slip10Node) DerivePath(indices ...uint32) (*Slip10Node, error) {
	node := n
	var err error
	for _, idx := range indices {
		node, err = node.DeriveHardened(idx)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (n *Slip10Node) PrivateKeyEd25519() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(n.Key)
}

func (n *Slip10Node) PublicKey() ([]byte, error) {
	priv := ed25519.NewKeyFromSeed(n.Key)
	pub := priv.Public().(ed25519.PublicKey)
	if len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("unexpected ed25519 public key size: %d", len(pub))
	}
	return append([]byte(nil), pub...), nil
}
