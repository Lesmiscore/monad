package database

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/wakiyamap/monad/chaincfg"
	"github.com/wakiyamap/monad/wire"
	"github.com/wakiyamap/monautil"
)

const (
	// userCheckpointDbNamePrefix is the prefix for the monad usercheckpoint database.
	userCheckpointDbNamePrefix = "usercheckpoints"

	// volatileCheckpointDbNamePrefix is the prefix for the monad volatilecheckpoint database.
	volatileCheckpointDbNamePrefix = "volatilecheckpoints"

	defaultDbType = "leveldb"
)

var (
	monadHomeDir    = monautil.AppDataDir("monad", false)
	defaultDataDir  = filepath.Join(monadHomeDir, "data")
	activeNetParams = &chaincfg.MainNetParams
	testnet         = flag.Bool("testnet", false, "operate on the testnet Bitcoin network")
	regtest         = flag.Bool("regtest", false, "operate on the regtest Bitcoin network")
	simnet          = flag.Bool("simnet", false, "operate on the simnet Bitcoin network")
)

type UserCheckpoint struct {
	Ucdb *leveldb.DB
}

var instance *UserCheckpoint
var once sync.Once

// netName returns the name used when referring to a bitcoin network.  At the
// time of writing, monad currently places blocks for testnet version 3 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory name
// as "testnet" when the passed active network matches wire.TestNet4.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet4" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
func netName(chainParams *chaincfg.Params) string {
	switch chainParams.Net {
	case wire.TestNet4:
		return "testnet"
	default:
		return chainParams.Name
	}
}

// open usercheckpointDB. Basically it is called only at startup.
func (uc *UserCheckpoint) OpenDB() error {
	if uc.Ucdb != nil {
		return nil
	}

	var err error
	dbpath := GetUserCheckpointDbPath()
	uc.Ucdb, err = leveldb.OpenFile(dbpath, nil)
	return err
}

// Basically it is called only at the end.
func (uc *UserCheckpoint) CloseDB() {
	if uc.Ucdb == nil {
		return
	}
	uc.Ucdb.Close()
	uc.Ucdb = nil
}

func (uc *UserCheckpoint) Add(height int64, hash string) {
	_ = uc.Ucdb.Put([]byte(fmt.Sprintf("%020d", height)), []byte(hash), nil)
}

func (uc *UserCheckpoint) Delete(height int64) {
	_ = uc.Ucdb.Delete([]byte(fmt.Sprintf("%020d", height)), nil)
}

func (uc *UserCheckpoint) GetMaxCheckpointHeight() (height int64) {
	height = 0
	iter := uc.Ucdb.NewIterator(nil, nil)
	iter.Last()

	if !iter.Valid() {
		return height
	}

	height, _ = strconv.ParseInt(string(iter.Key()), 10, 64)
	iter.Release()
	return height
}

func GetUserCheckpointDbInstance() *UserCheckpoint {
	once.Do(func() {
		time.Sleep(1 * time.Second)
		instance = &UserCheckpoint{nil}
	})
	return instance
}

func GetUserCheckpointDbPath() (dbPath string) {
	flag.Parse()
	if *testnet {
		activeNetParams = &chaincfg.TestNet4Params
	}
	if *regtest {
		activeNetParams = &chaincfg.RegressionNetParams
	}
	if *simnet {
		activeNetParams = &chaincfg.SimNetParams
	}
	dbName := userCheckpointDbNamePrefix + "_" + defaultDbType
	dbPath = filepath.Join(defaultDataDir, netName(activeNetParams), dbName)

	return dbPath
}

type VolatileCheckpoint struct {
	Vcdb *leveldb.DB
}

var vinstance *VolatileCheckpoint
var vonce sync.Once

// open volatilecheckpointDB. Basically it is called only at startup.
func (vc *VolatileCheckpoint) OpenDB() error {
	if vc.Vcdb != nil {
		return nil
	}

	var err error
	dbpath := GetVolatileCheckpointDbPath()
	vc.Vcdb, err = leveldb.OpenFile(dbpath, nil)
	return err
}

// close volatilecheckpointDB. Basically it is called only at the end.
func (vc *VolatileCheckpoint) CloseDB() {
	if vc.Vcdb == nil {
		return
	}
	vc.Vcdb.Close()
	vc.Vcdb = nil
}

func (vc *VolatileCheckpoint) Set(height int64, hash string) {
	_ = vc.Vcdb.Put([]byte(fmt.Sprintf("%020d", height)), []byte(hash), nil)
}

func (vc *VolatileCheckpoint) ClearDB() {
	iter := vc.Vcdb.NewIterator(nil, nil)
	for iter.Next() {
		err := vc.Vcdb.Delete([]byte(string(iter.Key())), nil)
		if err != nil {
			break
		}
	}
	iter.Release()
}

func GetVolatileCheckpointDbInstance() *VolatileCheckpoint {
	vonce.Do(func() {
		time.Sleep(1 * time.Second)
		vinstance = &VolatileCheckpoint{nil}
	})
	return vinstance
}

func GetVolatileCheckpointDbPath() (dbPath string) {
	flag.Parse()
	if *testnet {
		activeNetParams = &chaincfg.TestNet4Params
	}
	if *regtest {
		activeNetParams = &chaincfg.RegressionNetParams
	}
	if *simnet {
		activeNetParams = &chaincfg.SimNetParams
	}
	dbName := volatileCheckpointDbNamePrefix + "_" + defaultDbType
	dbPath = filepath.Join(defaultDataDir, netName(activeNetParams), dbName)

	return dbPath
}
