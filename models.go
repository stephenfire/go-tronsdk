package go_tronsdk

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

type WitnessPerm struct {
	OwnerAddr   address.Address
	WitnessAddr address.Address
}

func (p *WitnessPerm) WitnessPermAddr() address.Address {
	if len(p.WitnessAddr) > 0 {
		return p.WitnessAddr
	}
	return p.OwnerAddr
}

func BytesToPrivateKey(priv []byte) (*ecdsa.PrivateKey, error) {
	p := new(ecdsa.PrivateKey)
	bitCurve := secp256k1.S256()
	p.PublicKey.Curve = bitCurve
	p.D = new(big.Int).SetBytes(priv)
	if p.D.Cmp(bitCurve.N) >= 0 {
		return nil, errors.New("invalide private key, >=N")
	}
	if p.D.Sign() <= 0 {
		return nil, errors.New("invalid private key, zero or negative")
	}
	p.PublicKey.X, p.PublicKey.Y = p.PublicKey.Curve.ScalarBaseMult(priv)
	if p.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return p, nil
}

func ByteToPublicKey(pub []byte) (*ecdsa.PublicKey, error) {
	p := new(ecdsa.PublicKey)
	x, y := elliptic.Unmarshal(secp256k1.S256(), pub)
	if x == nil {
		return nil, errors.New("invalid secp256k1 public key")
	}
	p.Curve = secp256k1.S256()
	p.X = x
	p.Y = y
	return p, nil
}
