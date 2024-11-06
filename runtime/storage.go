/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runtime

import (
	"fmt"
	"runtime"
	"sort"

	"github.com/fxamacker/cbor/v2"
	"github.com/onflow/atree"

	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/common/orderedmap"
	"github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/stdlib"
)

const (
	StorageDomainContract = "contract"
	AccountStorageKey     = "stored"
)

type Storage struct {
	*atree.PersistentSlabStorage

	// cachedDomainStorageMaps is a cache of domain storage maps.
	// Key is StorageKey{address, domain} and value is domain storage map.
	cachedDomainStorageMaps map[interpreter.StorageKey]*interpreter.DomainStorageMap

	// contractUpdates is a cache of contract updates.
	// Key is StorageKey{contract_address, contract_name} and value is contract composite value.
	contractUpdates *orderedmap.OrderedMap[interpreter.StorageKey, *interpreter.CompositeValue]

	Ledger atree.Ledger

	memoryGauge common.MemoryGauge

	accountStorageV1 *AccountStorageV1
	accountStorageV2 *AccountStorageV2
}

var _ atree.SlabStorage = &Storage{}
var _ interpreter.Storage = &Storage{}

func NewStorage(ledger atree.Ledger, memoryGauge common.MemoryGauge) *Storage {
	decodeStorable := func(
		decoder *cbor.StreamDecoder,
		slabID atree.SlabID,
		inlinedExtraData []atree.ExtraData,
	) (
		atree.Storable,
		error,
	) {
		return interpreter.DecodeStorable(
			decoder,
			slabID,
			inlinedExtraData,
			memoryGauge,
		)
	}

	decodeTypeInfo := func(decoder *cbor.StreamDecoder) (atree.TypeInfo, error) {
		return interpreter.DecodeTypeInfo(decoder, memoryGauge)
	}

	ledgerStorage := atree.NewLedgerBaseStorage(ledger)
	persistentSlabStorage := atree.NewPersistentSlabStorage(
		ledgerStorage,
		interpreter.CBOREncMode,
		interpreter.CBORDecMode,
		decodeStorable,
		decodeTypeInfo,
	)

	accountStorageV1 := NewAccountStorageV1(
		ledger,
		persistentSlabStorage,
		memoryGauge,
	)
	accountStorageV2 := NewAccountStorageV2(
		ledger,
		persistentSlabStorage,
		memoryGauge,
	)

	return &Storage{
		Ledger:                ledger,
		PersistentSlabStorage: persistentSlabStorage,
		memoryGauge:           memoryGauge,
		accountStorageV1:      accountStorageV1,
		accountStorageV2:      accountStorageV2,
	}
}

const storageIndexLength = 8

// GetDomainStorageMap returns existing or new domain storage map for the given account and domain.
func (s *Storage) GetDomainStorageMap(
	inter *interpreter.Interpreter,
	address common.Address,
	domain string,
	createIfNotExists bool,
) (
	domainStorageMap *interpreter.DomainStorageMap,
) {
	// Get cached domain storage map if it exists.

	domainStorageKey := interpreter.NewStorageKey(s.memoryGauge, address, domain)

	if s.cachedDomainStorageMaps != nil {
		domainStorageMap = s.cachedDomainStorageMaps[domainStorageKey]
		if domainStorageMap != nil {
			return domainStorageMap
		}
	}

	defer func() {
		// Cache domain storage map
		if domainStorageMap != nil {
			s.cacheDomainStorageMap(
				domainStorageKey,
				domainStorageMap,
			)
		}
	}()

	// TODO:
	const useV2 = true

	if useV2 {
		domainStorageMap = s.accountStorageV2.GetDomainStorageMap(
			inter,
			address,
			domain,
			createIfNotExists,
		)
	} else {
		domainStorageMap = s.accountStorageV1.GetDomainStorageMap(
			address,
			domain,
			createIfNotExists,
		)
	}
	return domainStorageMap
}

func (s *Storage) cacheDomainStorageMap(
	domainStorageKey interpreter.StorageKey,
	domainStorageMap *interpreter.DomainStorageMap,
) {
	if s.cachedDomainStorageMaps == nil {
		s.cachedDomainStorageMaps = map[interpreter.StorageKey]*interpreter.DomainStorageMap{}
	}

	s.cachedDomainStorageMaps[domainStorageKey] = domainStorageMap
}

// getSlabIndexFromRegisterValue returns register value as atree.SlabIndex.
// This function returns error if
// - underlying ledger panics, or
// - underlying ledger returns error when retrieving ledger value, or
// - retrieved ledger value is invalid (for atree.SlabIndex).
func getSlabIndexFromRegisterValue(
	ledger atree.Ledger,
	address common.Address,
	key []byte,
) (atree.SlabIndex, bool, error) {
	var data []byte
	var err error
	errors.WrapPanic(func() {
		data, err = ledger.GetValue(address[:], key)
	})
	if err != nil {
		return atree.SlabIndex{}, false, interpreter.WrappedExternalError(err)
	}

	dataLength := len(data)

	if dataLength == 0 {
		return atree.SlabIndex{}, false, nil
	}

	isStorageIndex := dataLength == storageIndexLength
	if !isStorageIndex {
		// Invalid data in register

		// TODO: add dedicated error type?
		return atree.SlabIndex{}, false, errors.NewUnexpectedError(
			"invalid storage index for storage map of account '%x': expected length %d, got %d",
			address[:], storageIndexLength, dataLength,
		)
	}

	return atree.SlabIndex(data), true, nil
}

func (s *Storage) recordContractUpdate(
	location common.AddressLocation,
	contractValue *interpreter.CompositeValue,
) {
	key := interpreter.NewStorageKey(s.memoryGauge, location.Address, location.Name)

	// NOTE: do NOT delete the map entry,
	// otherwise the removal write is lost

	if s.contractUpdates == nil {
		s.contractUpdates = &orderedmap.OrderedMap[interpreter.StorageKey, *interpreter.CompositeValue]{}
	}
	s.contractUpdates.Set(key, contractValue)
}

func (s *Storage) contractUpdateRecorded(
	location common.AddressLocation,
) bool {
	if s.contractUpdates == nil {
		return false
	}

	key := interpreter.NewStorageKey(s.memoryGauge, location.Address, location.Name)
	return s.contractUpdates.Contains(key)
}

type ContractUpdate struct {
	ContractValue *interpreter.CompositeValue
	Key           interpreter.StorageKey
}

func SortContractUpdates(updates []ContractUpdate) {
	sort.Slice(updates, func(i, j int) bool {
		a := updates[i].Key
		b := updates[j].Key
		return a.IsLess(b)
	})
}

// commitContractUpdates writes the contract updates to storage.
// The contract updates were delayed so they are not observable during execution.
func (s *Storage) commitContractUpdates(inter *interpreter.Interpreter) {
	if s.contractUpdates == nil {
		return
	}

	for pair := s.contractUpdates.Oldest(); pair != nil; pair = pair.Next() {
		s.writeContractUpdate(inter, pair.Key, pair.Value)
	}
}

func (s *Storage) writeContractUpdate(
	inter *interpreter.Interpreter,
	key interpreter.StorageKey,
	contractValue *interpreter.CompositeValue,
) {
	storageMap := s.GetDomainStorageMap(inter, key.Address, StorageDomainContract, true)
	// NOTE: pass nil instead of allocating a Value-typed  interface that points to nil
	storageMapKey := interpreter.StringStorageMapKey(key.Key)
	if contractValue == nil {
		storageMap.WriteValue(inter, storageMapKey, nil)
	} else {
		storageMap.WriteValue(inter, storageMapKey, contractValue)
	}
}

// Commit serializes/saves all values in the readCache in storage (through the runtime interface).
func (s *Storage) Commit(inter *interpreter.Interpreter, commitContractUpdates bool) error {
	return s.commit(inter, commitContractUpdates, true)
}

// NondeterministicCommit serializes and commits all values in the deltas storage
// in nondeterministic order.  This function is used when commit ordering isn't
// required (e.g. migration programs).
func (s *Storage) NondeterministicCommit(inter *interpreter.Interpreter, commitContractUpdates bool) error {
	return s.commit(inter, commitContractUpdates, false)
}

func (s *Storage) commit(inter *interpreter.Interpreter, commitContractUpdates bool, deterministic bool) error {

	if commitContractUpdates {
		s.commitContractUpdates(inter)
	}

	err := s.accountStorageV1.commit()
	if err != nil {
		return err
	}

	err = s.accountStorageV2.commit()
	if err != nil {
		return err
	}

	// Commit the underlying slab storage's writes

	size := s.PersistentSlabStorage.DeltasSizeWithoutTempAddresses()
	if size > 0 {
		inter.ReportComputation(common.ComputationKindEncodeValue, uint(size))
		usage := common.NewBytesMemoryUsage(int(size))
		common.UseMemory(s.memoryGauge, usage)
	}

	deltas := s.PersistentSlabStorage.DeltasWithoutTempAddresses()
	common.UseMemory(s.memoryGauge, common.NewAtreeEncodedSlabMemoryUsage(deltas))

	// TODO: report encoding metric for all encoded slabs
	if deterministic {
		return s.PersistentSlabStorage.FastCommit(runtime.NumCPU())
	} else {
		return s.PersistentSlabStorage.NondeterministicFastCommit(runtime.NumCPU())
	}
}

func commitSlabIndices(
	slabIndices *orderedmap.OrderedMap[interpreter.StorageKey, atree.SlabIndex],
	ledger atree.Ledger,
) error {
	for pair := slabIndices.Oldest(); pair != nil; pair = pair.Next() {
		var err error
		errors.WrapPanic(func() {
			err = ledger.SetValue(
				pair.Key.Address[:],
				[]byte(pair.Key.Key),
				pair.Value[:],
			)
		})
		if err != nil {
			return interpreter.WrappedExternalError(err)
		}
	}

	return nil
}

// TODO:
//func (s *Storage) migrateAccounts(inter *interpreter.Interpreter) error {
//	// Predicate function allows migration for accounts with write ops.
//	migrateAccountPred := func(address common.Address) bool {
//		return s.PersistentSlabStorage.HasUnsavedChanges(atree.Address(address))
//	}
//
//	// getDomainStorageMap function returns cached domain storage map if it is available
//	// before loading domain storage map from storage.
//	// This is necessary to migrate uncommitted (new) but cached domain storage map.
//	getDomainStorageMap := func(
//		ledger atree.Ledger,
//		storage atree.SlabStorage,
//		address common.Address,
//		domain string,
//	) (*interpreter.DomainStorageMap, error) {
//		domainStorageKey := interpreter.NewStorageKey(s.memoryGauge, address, domain)
//
//		// Get cached domain storage map if available.
//		domainStorageMap := s.cachedDomainStorageMaps[domainStorageKey]
//
//		if domainStorageMap != nil {
//			return domainStorageMap, nil
//		}
//
//		return getDomainStorageMapFromLegacyDomainRegister(ledger, storage, address, domain)
//	}
//
//	migrator := NewDomainRegisterMigration(s.Ledger, s.PersistentSlabStorage, inter, s.memoryGauge)
//	migrator.SetGetDomainStorageMapFunc(getDomainStorageMap)
//
//	migratedAccounts, err := migrator.MigrateAccounts(s.unmigratedAccounts, migrateAccountPred)
//	if err != nil {
//		return err
//	}
//
//	if migratedAccounts == nil {
//		return nil
//	}
//
//	// Update internal state with migrated accounts
//	for pair := migratedAccounts.Oldest(); pair != nil; pair = pair.Next() {
//		address := pair.Key
//		accountStorageMap := pair.Value
//
//		// Cache migrated account storage map
//		accountStorageKey := interpreter.NewStorageKey(s.memoryGauge, address, AccountStorageKey)
//		s.cachedAccountStorageMaps[accountStorageKey] = accountStorageMap
//
//		// Remove migrated accounts from unmigratedAccounts
//		s.unmigratedAccounts.Delete(address)
//	}
//
//	return nil
//}

func (s *Storage) CheckHealth() error {
	// TODO:

	//// Check slab storage health
	//rootSlabIDs, err := atree.CheckStorageHealth(s, -1)
	//if err != nil {
	//	return err
	//}
	//
	//// Find account / non-temporary root slab IDs
	//
	//accountRootSlabIDs := make(map[atree.SlabID]struct{}, len(rootSlabIDs))
	//
	//// NOTE: map range is safe, as it creates a subset
	//for rootSlabID := range rootSlabIDs { //nolint:maprange
	//	if rootSlabID.HasTempAddress() {
	//		continue
	//	}
	//
	//	accountRootSlabIDs[rootSlabID] = struct{}{}
	//}
	//
	//// Check that account storage maps and unmigrated domain storage maps
	//// match returned root slabs from atree.CheckStorageHealth.
	//
	//var storageMapStorageIDs []atree.SlabID
	//
	//// Get cached account storage map slab IDs.
	//for _, storageMap := range s.cachedAccountStorageMaps { //nolint:maprange
	//	storageMapStorageIDs = append(
	//		storageMapStorageIDs,
	//		storageMap.SlabID(),
	//	)
	//}
	//
	//// Get cached unmigrated domain storage map slab IDs
	//for storageKey, storageMap := range s.cachedDomainStorageMaps { //nolint:maprange
	//	address := storageKey.Address
	//
	//	if s.unmigratedAccounts != nil &&
	//		s.unmigratedAccounts.Contains(address) {
	//
	//		domainValueID := storageMap.ValueID()
	//
	//		slabID := atree.NewSlabID(
	//			atree.Address(address),
	//			atree.SlabIndex(domainValueID[8:]),
	//		)
	//
	//		storageMapStorageIDs = append(
	//			storageMapStorageIDs,
	//			slabID,
	//		)
	//	}
	//}
	//
	//sort.Slice(storageMapStorageIDs, func(i, j int) bool {
	//	a := storageMapStorageIDs[i]
	//	b := storageMapStorageIDs[j]
	//	return a.Compare(b) < 0
	//})
	//
	//found := map[atree.SlabID]struct{}{}
	//
	//for _, storageMapStorageID := range storageMapStorageIDs {
	//	if _, ok := accountRootSlabIDs[storageMapStorageID]; !ok {
	//		return errors.NewUnexpectedError("account storage map (and unmigrated domain storage map) points to non-root slab %s", storageMapStorageID)
	//	}
	//
	//	found[storageMapStorageID] = struct{}{}
	//}
	//
	//// Check that all slabs in slab storage
	//// are referenced by storables in account storage.
	//// If a slab is not referenced, it is garbage.
	//
	//if len(accountRootSlabIDs) > len(found) {
	//	var unreferencedRootSlabIDs []atree.SlabID
	//
	//	for accountRootSlabID := range accountRootSlabIDs { //nolint:maprange
	//		if _, ok := found[accountRootSlabID]; ok {
	//			continue
	//		}
	//
	//		unreferencedRootSlabIDs = append(
	//			unreferencedRootSlabIDs,
	//			accountRootSlabID,
	//		)
	//	}
	//
	//	sort.Slice(unreferencedRootSlabIDs, func(i, j int) bool {
	//		a := unreferencedRootSlabIDs[i]
	//		b := unreferencedRootSlabIDs[j]
	//		return a.Compare(b) < 0
	//	})
	//
	//	return UnreferencedRootSlabsError{
	//		UnreferencedRootSlabIDs: unreferencedRootSlabIDs,
	//	}
	//}

	return nil
}

type UnreferencedRootSlabsError struct {
	UnreferencedRootSlabIDs []atree.SlabID
}

var _ errors.InternalError = UnreferencedRootSlabsError{}

func (UnreferencedRootSlabsError) IsInternalError() {}

func (e UnreferencedRootSlabsError) Error() string {
	return fmt.Sprintf("slabs not referenced: %s", e.UnreferencedRootSlabIDs)
}

var AccountDomains = []string{
	common.PathDomainStorage.Identifier(),
	common.PathDomainPrivate.Identifier(),
	common.PathDomainPublic.Identifier(),
	StorageDomainContract,
	stdlib.InboxStorageDomain,
	stdlib.CapabilityControllerStorageDomain,
	stdlib.CapabilityControllerTagStorageDomain,
	stdlib.PathCapabilityStorageDomain,
	stdlib.AccountCapabilityStorageDomain,
}

var accountDomainsSet = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, domain := range AccountDomains {
		m[domain] = struct{}{}
	}
	return m
}()
