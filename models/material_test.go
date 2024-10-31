package models

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evolutionlandorg/evo-backend/config"
	"github.com/evolutionlandorg/evo-backend/util"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type EthAddress struct {
	PrivateKey string
	Address    string
}

func (e EthAddress) Signature(data []byte) (string, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(e.PrivateKey, "0x"))
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

func NewEthAddress() (EthAddress, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return EthAddress{}, errors.Wrap(err, "generate key failed")
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return EthAddress{
		PrivateKey: hexutil.Encode(privateKeyBytes),
		Address:    address,
	}, nil
}

func Test_materialTakeBack(t *testing.T) {
	config.InitApplication()
	util.Panic(util.InitMysql())
	util.Panic(util.InitRedis())

	// 余额足以抵扣
	var (
		ctx        = context.Background()
		chain      = EthChain
		material   = "MS-1"
		materialId = 4
	)
	eth, err := NewEthAddress()
	assert.NoError(t, err)
	uId, err := NewMember(ctx, chain, eth.Address, eth.Address, 0)
	assert.NoError(t, err)
	defer func() {
		util.WithContextDb(ctx).Where("id = ?", uId).Limit(1).Delete(new(Member))
	}()
	member := GetMember(ctx, int(uId))

	account := member.TouchAccount(ctx, material, eth.Address, chain, true)
	defer func() {
		util.WithContextDb(ctx).Where("member_id = ?", uId).Delete(new(Account))
	}()

	parseTx := func(f func(txn *util.GormDB)) {
		txn := util.DbBegin(ctx)
		f(txn)
		txn.DbCommit()
	}

	// 余额刚好够抵扣
	amount := decimal.NewFromFloat(20)
	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, account.AddBalance(txn, amount, ReasonDungeonClearance))
	})

	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, materialTakeBack(txn, "fake", eth.Address, chain, materialId, amount))
	})
	account = member.TouchAccount(ctx, material, eth.Address, chain, true)
	assert.True(t, account.Balance.IsZero())

	// 余额比amount多一点
	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, account.AddBalance(txn, amount, ReasonDungeonClearance))
	})

	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, materialTakeBack(txn, "fake", eth.Address, chain, materialId, amount.Sub(decimal.NewFromInt32(10))))
	})
	account = member.TouchAccount(ctx, material, eth.Address, chain, true)
	assert.Equal(t, account.Balance.String(), amount.Sub(decimal.NewFromInt32(10)).String())

	// 余额比amount少, 但是多chain余额加起来多
	crabAccount := member.TouchAccount(ctx, material, eth.Address, CrabChain, true)
	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, crabAccount.AddBalance(txn, decimal.NewFromInt32(11), ReasonDungeonClearance))
	})

	parseTx(func(txn *util.GormDB) {
		assert.NoError(t, materialTakeBack(txn, "fake", eth.Address, chain, materialId, amount))
	})
	crabAccount = member.TouchAccount(ctx, material, eth.Address, CrabChain, true)
	assert.Equal(t, crabAccount.Balance.String(), decimal.NewFromInt32(1).String())

	account = member.TouchAccount(ctx, material, eth.Address, chain, true)
	assert.Equal(t, account.Balance.String(), decimal.NewFromInt32(0).String())
}
