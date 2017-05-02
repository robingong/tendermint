package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/config/tendermint_test"
	dbm "github.com/tendermint/tmlibs/db"
)

func ResetConfig(name string) *cfg.Config {
	viperConfig := tendermint_test.ResetConfig(name)
	config := new(struct {
		cfg.Config `mapstructure:",squash"`
	})
	if err := viperConfig.Unmarshal(config); err != nil {
		panic(err)
	}
	return &config.Config
}

func TestStateCopyEquals(t *testing.T) {
	config := ResetConfig("state_")
	// Get State db
	stateDB := dbm.NewDB("state", config.DBBackend, config.DBDir)
	state := GetState(stateDB, config.GenesisFile)

	stateCopy := state.Copy()

	if !state.Equals(stateCopy) {
		t.Fatal("expected state and its copy to be identical. got %v\n expected %v\n", stateCopy, state)
	}

	stateCopy.LastBlockHeight += 1

	if state.Equals(stateCopy) {
		t.Fatal("expected states to be different. got same %v", state)
	}
}

func TestStateSaveLoad(t *testing.T) {
	config := ResetConfig("state_")
	// Get State db
	stateDB := dbm.NewDB("state", config.DBBackend, config.DBDir)
	state := GetState(stateDB, config.GenesisFile)

	state.LastBlockHeight += 1
	state.Save()

	loadedState := LoadState(stateDB)
	if !state.Equals(loadedState) {
		t.Fatal("expected state and its copy to be identical. got %v\n expected %v\n", loadedState, state)
	}
}

func TestABCIResponsesSaveLoad(t *testing.T) {
	assert := assert.New(t)

	config := ResetConfig("state_")
	stateDB := dbm.NewDB("state", config.DBBackend, config.DBDir)
	state := GetState(stateDB, config.GenesisFile)

	state.LastBlockHeight += 1

	// build mock responses
	block := makeBlock(2, state)
	abciResponses := NewABCIResponses(block)
	abciResponses.DeliverTx[0] = &abci.ResponseDeliverTx{Data: []byte("foo")}
	abciResponses.DeliverTx[1] = &abci.ResponseDeliverTx{Data: []byte("bar"), Log: "ok"}
	abciResponses.EndBlock = abci.ResponseEndBlock{Diffs: []*abci.Validator{
		{
			PubKey: crypto.GenPrivKeyEd25519().PubKey().Bytes(),
			Power:  10,
		},
	}}
	abciResponses.txs = nil

	state.SaveABCIResponses(abciResponses)
	abciResponses2 := state.LoadABCIResponses()
	assert.Equal(abciResponses, abciResponses2, fmt.Sprintf("ABCIResponses don't match: Got %v, Expected %v", abciResponses2, abciResponses))
}
