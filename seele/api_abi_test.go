/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */
package seele

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/seeleteam/go-seele/common"

	"github.com/stretchr/testify/assert"
)

const (
	SimpleStorageABI  = "[{\"constant\":false,\"inputs\":[{\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"
	RemixGetPayload   = "0x6d4ce63c"
	RemixSet23Payload = "0x60fe47b10000000000000000000000000000000000000000000000000000000000000017"
)

func Test_GeneratePayload(t *testing.T) {
	dbPath := filepath.Join(common.GetTempFolder(), ".GeneratePayload")
	if common.FileOrFolderExists(dbPath) {
		os.RemoveAll(dbPath)
	}
	api := newTestAPI(t, dbPath)

	// Get method test
	payload1, err1 := api.GeneratePayload(SimpleStorageABI, "get")
	assert.NoError(t, err1)
	assert.Equal(t, payload1, RemixGetPayload)

	// Set method test
	payload2, err2 := api.GeneratePayload(SimpleStorageABI, "set", big.NewInt(23))
	assert.NoError(t, err2)
	assert.Equal(t, payload2, RemixSet23Payload)

	// Invalid method test
	payload3, err3 := api.GeneratePayload(SimpleStorageABI, "add", big.NewInt(23))
	assert.Error(t, err3)
	assert.Empty(t, payload3)

	// Invalid parameter type test
	payload4, err4 := api.GeneratePayload(SimpleStorageABI, "set", 23)
	assert.Error(t, err4)
	assert.Empty(t, payload4)

	// Invalid abiJSON string test
	payload5, err5 := api.GeneratePayload("SimpleStorageABI:asdf", "set", 23)
	assert.Error(t, err5)
	assert.Empty(t, payload5)
}

func Test_GetAPI(t *testing.T) {
	dbPath := filepath.Join(common.GetTempFolder(), ".GetAPI")
	if common.FileOrFolderExists(dbPath) {
		os.RemoveAll(dbPath)
	}
	api := newTestAPI(t, dbPath)

	from, err := saveABI(api)
	assert.NoError(t, err)

	// Correctness test
	abiObj1, err1 := api.GetABI(from)
	assert.NoError(t, err1)
	assert.Equal(t, abiObj1, SimpleStorageABI)

	// Invalid address test
	abiObj2, err2 := api.GetABI(common.EmptyAddress)
	assert.Error(t, err2)
	assert.Equal(t, abiObj2, "")
}

func saveABI(api *PublicSeeleAPI) (common.Address, error) {
	statedb, err := api.s.chain.GetCurrentState()
	if err != nil {
		return common.EmptyAddress, err
	}

	from := getFromAddress(statedb)
	statedb.SetData(from, KeyABIHash, []byte(SimpleStorageABI))

	// save the statedb
	batch := api.s.accountStateDB.NewBatch()
	block := api.s.chain.CurrentBlock()
	block.Header.StateHash, _ = statedb.Commit(batch)
	block.Header.Height++
	block.Header.PreviousBlockHash = block.HeaderHash
	block.HeaderHash = block.Header.Hash()
	api.s.chain.GetStore().PutBlock(block, big.NewInt(1), true)
	batch.Commit()
	return from, nil
}
