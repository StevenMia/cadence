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

package interpreter

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/pretty"
	"github.com/onflow/cadence/sema"
)

// unsupportedOperation

type unsupportedOperation struct {
	kind      common.OperationKind
	operation ast.Operation
	ast.Range
}

var _ errors.InternalError = &unsupportedOperation{}

func (*unsupportedOperation) IsInternalError() {}

func (e *unsupportedOperation) Error() string {
	return fmt.Sprintf(
		"%s cannot evaluate unsupported %s operation: %s",
		errors.InternalErrorMessagePrefix,
		e.kind.Name(),
		e.operation.Symbol(),
	)
}

// Error is the containing type for all errors produced by the interpreter.
type Error struct {
	Err        error
	Location   common.Location
	StackTrace []Invocation
}

func (e Error) Unwrap() error {
	return e.Err
}

func (e Error) Error() string {
	var sb strings.Builder
	sb.WriteString("Execution failed:\n")
	printErr := pretty.NewErrorPrettyPrinter(&sb, false).
		PrettyPrintError(e.Err, e.Location, map[common.Location][]byte{})
	if printErr != nil {
		panic(printErr)
	}
	return sb.String()
}

func (e Error) ChildErrors() []error {
	errs := make([]error, 0, 1+len(e.StackTrace))

	for _, invocation := range e.StackTrace {
		locationRange := invocation.LocationRange
		if locationRange.Location == nil {
			continue
		}

		errs = append(
			errs,
			StackTraceError{
				LocationRange: locationRange,
			},
		)
	}

	return append(errs, e.Err)
}

func (e Error) ImportLocation() common.Location {
	return e.Location
}

type StackTraceError struct {
	LocationRange
}

func (e StackTraceError) Error() string {
	return ""
}

func (e StackTraceError) Prefix() string {
	return ""
}

func (e StackTraceError) ImportLocation() common.Location {
	return e.Location
}

// PositionedError wraps an unpositioned error with position info
type PositionedError struct {
	Err error
	ast.Range
}

var _ errors.UserError = PositionedError{}

func (PositionedError) IsUserError() {}

func (e PositionedError) Unwrap() error {
	return e.Err
}

func (e PositionedError) Error() string {
	return e.Err.Error()
}

// NotDeclaredError

type NotDeclaredError struct {
	Name         string
	ExpectedKind common.DeclarationKind
}

var _ errors.UserError = NotDeclaredError{}
var _ errors.SecondaryError = NotDeclaredError{}

func (NotDeclaredError) IsUserError() {}

func (e NotDeclaredError) Error() string {
	return fmt.Sprintf(
		"cannot find %s in this scope: `%s`",
		e.ExpectedKind.Name(),
		e.Name,
	)
}

func (e NotDeclaredError) SecondaryError() string {
	return "not found in this scope"
}

// NotInvokableError

type NotInvokableError struct {
	Value Value
}

var _ errors.UserError = NotInvokableError{}

func (NotInvokableError) IsUserError() {}

func (e NotInvokableError) Error() string {
	return fmt.Sprintf("cannot call value: %#+v", e.Value)
}

// ArgumentCountError

type ArgumentCountError struct {
	ParameterCount int
	ArgumentCount  int
}

var _ errors.UserError = ArgumentCountError{}

func (ArgumentCountError) IsUserError() {}

func (e ArgumentCountError) Error() string {
	return fmt.Sprintf(
		"incorrect number of arguments: expected %d, got %d",
		e.ParameterCount,
		e.ArgumentCount,
	)
}

// TransactionNotDeclaredError

type TransactionNotDeclaredError struct {
	Index int
}

var _ errors.UserError = TransactionNotDeclaredError{}

func (TransactionNotDeclaredError) IsUserError() {}

func (e TransactionNotDeclaredError) Error() string {
	return fmt.Sprintf(
		"cannot find transaction with index %d in this scope",
		e.Index,
	)
}

// ConditionError

type ConditionError struct {
	LocationRange
	Message       string
	ConditionKind ast.ConditionKind
}

var _ errors.UserError = ConditionError{}

func (ConditionError) IsUserError() {}

func (e ConditionError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("%s failed", e.ConditionKind.Name())
	}
	return fmt.Sprintf("%s failed: %s", e.ConditionKind.Name(), e.Message)
}

// RedeclarationError

type RedeclarationError struct {
	Name string
}

var _ errors.UserError = RedeclarationError{}

func (RedeclarationError) IsUserError() {}

func (e RedeclarationError) Error() string {
	return fmt.Sprintf("cannot redeclare: `%s` is already declared", e.Name)
}

// DereferenceError

type DereferenceError struct {
	Cause        string
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.UserError = DereferenceError{}
var _ errors.SecondaryError = DereferenceError{}

func (DereferenceError) IsUserError() {}

func (e DereferenceError) Error() string {
	return "dereference failed"
}

func (e DereferenceError) SecondaryError() string {
	if e.Cause != "" {
		return e.Cause
	}
	expected, actual := sema.ErrorMessageExpectedActualTypes(
		e.ExpectedType,
		e.ActualType,
	)

	return fmt.Sprintf(
		"type mismatch: expected `%s`, got `%s`",
		expected,
		actual,
	)
}

// OverflowError

type OverflowError struct {
	LocationRange
}

var _ errors.UserError = OverflowError{}

func (OverflowError) IsUserError() {}

func (e OverflowError) Error() string {
	return "overflow"
}

// UnderflowError

type UnderflowError struct {
	LocationRange
}

var _ errors.UserError = UnderflowError{}

func (UnderflowError) IsUserError() {}

func (e UnderflowError) Error() string {
	return "underflow"
}

// NegativeShiftError

type NegativeShiftError struct {
	LocationRange
}

var _ errors.UserError = NegativeShiftError{}

func (NegativeShiftError) IsUserError() {}

func (e NegativeShiftError) Error() string {
	return "negative shift"
}

// DivisionByZeroError

type DivisionByZeroError struct {
	LocationRange
}

var _ errors.UserError = DivisionByZeroError{}

func (DivisionByZeroError) IsUserError() {}

func (e DivisionByZeroError) Error() string {
	return "division by zero"
}

// InvalidatedResourceError
type InvalidatedResourceError struct {
	LocationRange
}

var _ errors.InternalError = InvalidatedResourceError{}

func (InvalidatedResourceError) IsInternalError() {}

func (e InvalidatedResourceError) Error() string {
	return fmt.Sprintf(
		"%s resource is invalidated and cannot be used anymore",
		errors.InternalErrorMessagePrefix,
	)
}

// DestroyedResourceError is the error which is reported
// when a user uses a destroyed resource through a reference
type DestroyedResourceError struct {
	LocationRange
}

var _ errors.UserError = DestroyedResourceError{}

func (DestroyedResourceError) IsUserError() {}

func (e DestroyedResourceError) Error() string {
	return "resource was destroyed and cannot be used anymore"
}

// ForceNilError
type ForceNilError struct {
	LocationRange
}

var _ errors.UserError = ForceNilError{}

func (ForceNilError) IsUserError() {}

func (e ForceNilError) Error() string {
	return "unexpectedly found nil while forcing an Optional value"
}

// ForceCastTypeMismatchError
type ForceCastTypeMismatchError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.UserError = ForceCastTypeMismatchError{}

func (ForceCastTypeMismatchError) IsUserError() {}

func (e ForceCastTypeMismatchError) Error() string {
	expected, actual := sema.ErrorMessageExpectedActualTypes(
		e.ExpectedType,
		e.ActualType,
	)

	return fmt.Sprintf(
		"failed to force-cast value: expected type `%s`, got `%s`",
		expected,
		actual,
	)
}

// TypeMismatchError
type TypeMismatchError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.UserError = TypeMismatchError{}

func (TypeMismatchError) IsUserError() {}

func (e TypeMismatchError) Error() string {
	expected, actual := sema.ErrorMessageExpectedActualTypes(
		e.ExpectedType,
		e.ActualType,
	)

	return fmt.Sprintf(
		"type mismatch: expected `%s`, got `%s`",
		expected,
		actual,
	)
}

// InvalidMemberReferenceError
type InvalidMemberReferenceError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.UserError = InvalidMemberReferenceError{}

func (InvalidMemberReferenceError) IsUserError() {}

func (e InvalidMemberReferenceError) Error() string {
	expected, actual := sema.ErrorMessageExpectedActualTypes(
		e.ExpectedType,
		e.ActualType,
	)

	return fmt.Sprintf(
		"cannot create reference: expected `%s`, got `%s`",
		expected,
		actual,
	)
}

// InvalidPathDomainError
type InvalidPathDomainError struct {
	LocationRange
	ExpectedDomains []common.PathDomain
	ActualDomain    common.PathDomain
}

var _ errors.UserError = InvalidPathDomainError{}
var _ errors.SecondaryError = InvalidPathDomainError{}

func (InvalidPathDomainError) IsUserError() {}

func (e InvalidPathDomainError) Error() string {
	return "invalid path domain"
}

func (e InvalidPathDomainError) SecondaryError() string {

	domainNames := make([]string, len(e.ExpectedDomains))

	for i, domain := range e.ExpectedDomains {
		domainNames[i] = domain.Identifier()
	}

	return fmt.Sprintf(
		"expected %s, got `%s`",
		common.EnumerateWords(domainNames, "or"),
		e.ActualDomain.Identifier(),
	)
}

// OverwriteError
type OverwriteError struct {
	LocationRange
	Path    PathValue
	Address AddressValue
}

var _ errors.UserError = OverwriteError{}

func (OverwriteError) IsUserError() {}

func (e OverwriteError) Error() string {
	return fmt.Sprintf(
		"failed to save object: path %s in account %s already stores an object",
		e.Path,
		e.Address,
	)
}

// ArrayIndexOutOfBoundsError
type ArrayIndexOutOfBoundsError struct {
	LocationRange
	Index int
	Size  int
}

var _ errors.UserError = ArrayIndexOutOfBoundsError{}

func (ArrayIndexOutOfBoundsError) IsUserError() {}

func (e ArrayIndexOutOfBoundsError) Error() string {
	return fmt.Sprintf(
		"array index out of bounds: %d, but size is %d",
		e.Index,
		e.Size,
	)
}

// ArraySliceIndicesError
type ArraySliceIndicesError struct {
	LocationRange
	FromIndex int
	UpToIndex int
	Size      int
}

var _ errors.UserError = ArraySliceIndicesError{}

func (ArraySliceIndicesError) IsUserError() {}

func (e ArraySliceIndicesError) Error() string {
	return fmt.Sprintf(
		"slice indices [%d:%d] are out of bounds (size %d)",
		e.FromIndex, e.UpToIndex, e.Size,
	)
}

// InvalidSliceIndexError is returned when a slice index is invalid, such as fromIndex > upToIndex
// This error can be returned even when fromIndex and upToIndex are both within bounds.
type InvalidSliceIndexError struct {
	LocationRange
	FromIndex int
	UpToIndex int
}

var _ errors.UserError = InvalidSliceIndexError{}

func (InvalidSliceIndexError) IsUserError() {}

func (e InvalidSliceIndexError) Error() string {
	return fmt.Sprintf("invalid slice index: %d > %d", e.FromIndex, e.UpToIndex)
}

// StringIndexOutOfBoundsError
type StringIndexOutOfBoundsError struct {
	LocationRange
	Index  int
	Length int
}

var _ errors.UserError = StringIndexOutOfBoundsError{}

func (StringIndexOutOfBoundsError) IsUserError() {}

func (e StringIndexOutOfBoundsError) Error() string {
	return fmt.Sprintf(
		"string index out of bounds: %d, but length is %d",
		e.Index,
		e.Length,
	)
}

// StringSliceIndicesError
type StringSliceIndicesError struct {
	LocationRange
	FromIndex int
	UpToIndex int
	Length    int
}

var _ errors.UserError = StringSliceIndicesError{}

func (StringSliceIndicesError) IsUserError() {}

func (e StringSliceIndicesError) Error() string {
	return fmt.Sprintf(
		"string slice indices [%d:%d] are out of bounds (length %d)",
		e.FromIndex, e.UpToIndex, e.Length,
	)
}

// EventEmissionUnavailableError
type EventEmissionUnavailableError struct {
	LocationRange
}

var _ errors.UserError = EventEmissionUnavailableError{}

func (EventEmissionUnavailableError) IsUserError() {}

func (e EventEmissionUnavailableError) Error() string {
	return "cannot emit event: event emission is unavailable in this configuration of Cadence"
}

// UUIDUnavailableError
type UUIDUnavailableError struct {
	LocationRange
}

var _ errors.UserError = UUIDUnavailableError{}

func (UUIDUnavailableError) IsUserError() {}

func (e UUIDUnavailableError) Error() string {
	return "cannot get UUID: UUID access is unavailable in this configuration of Cadence"
}

// TypeLoadingError
type TypeLoadingError struct {
	TypeID TypeID
}

var _ errors.UserError = TypeLoadingError{}

func (TypeLoadingError) IsUserError() {}

func (e TypeLoadingError) Error() string {
	return fmt.Sprintf("failed to load type: %s", e.TypeID)
}

// UseBeforeInitializationError
type UseBeforeInitializationError struct {
	LocationRange
	Name string
}

var _ errors.UserError = UseBeforeInitializationError{}

func (UseBeforeInitializationError) IsUserError() {}

func (e UseBeforeInitializationError) Error() string {
	return fmt.Sprintf("member `%s` is used before it has been initialized", e.Name)
}

// MemberAccessTypeError
type MemberAccessTypeError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.InternalError = MemberAccessTypeError{}

func (MemberAccessTypeError) IsInternalError() {}

func (e MemberAccessTypeError) Error() string {
	return fmt.Sprintf(
		"%s invalid member access: expected `%s`, got `%s`",
		errors.InternalErrorMessagePrefix,
		e.ExpectedType.QualifiedString(),
		e.ActualType.QualifiedString(),
	)
}

// ValueTransferTypeError
type ValueTransferTypeError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.InternalError = ValueTransferTypeError{}

func (ValueTransferTypeError) IsInternalError() {}

func (e ValueTransferTypeError) Error() string {
	expected, actual := sema.ErrorMessageExpectedActualTypes(
		e.ExpectedType,
		e.ActualType,
	)

	return fmt.Sprintf(
		"%s invalid transfer of value: expected `%s`, got `%s`",
		errors.InternalErrorMessagePrefix,
		expected,
		actual,
	)
}

// UnexpectedMappedEntitlementError
type UnexpectedMappedEntitlementError struct {
	Type sema.Type
	LocationRange
}

var _ errors.InternalError = UnexpectedMappedEntitlementError{}

func (UnexpectedMappedEntitlementError) IsInternalError() {}

func (e UnexpectedMappedEntitlementError) Error() string {
	return fmt.Sprintf(
		"%s invalid transfer of value: found an unexpected runtime mapped entitlement `%s`",
		errors.InternalErrorMessagePrefix,
		e.Type.QualifiedString(),
	)
}

// ResourceConstructionError
type ResourceConstructionError struct {
	CompositeType *sema.CompositeType
	LocationRange
}

var _ errors.InternalError = ResourceConstructionError{}

func (ResourceConstructionError) IsInternalError() {}

func (e ResourceConstructionError) Error() string {
	return fmt.Sprintf(
		"%s cannot create resource `%s`: outside of declaring location %s",
		errors.InternalErrorMessagePrefix,
		e.CompositeType.QualifiedString(),
		e.CompositeType.Location.String(),
	)
}

// ContainerMutationError
type ContainerMutationError struct {
	ExpectedType sema.Type
	ActualType   sema.Type
	LocationRange
}

var _ errors.UserError = ContainerMutationError{}

func (ContainerMutationError) IsUserError() {}

func (e ContainerMutationError) Error() string {
	return fmt.Sprintf(
		"invalid container update: expected a subtype of `%s`, found `%s`",
		e.ExpectedType.QualifiedString(),
		e.ActualType.QualifiedString(),
	)
}

// NonStorableValueError
type NonStorableValueError struct {
	Value Value
}

var _ errors.UserError = NonStorableValueError{}

func (NonStorableValueError) IsUserError() {}

func (e NonStorableValueError) Error() string {
	return "cannot store non-storable value"
}

// NonStorableStaticTypeError
type NonStorableStaticTypeError struct {
	Type sema.Type
}

var _ errors.UserError = NonStorableStaticTypeError{}

func (NonStorableStaticTypeError) IsUserError() {}

func (e NonStorableStaticTypeError) Error() string {
	return fmt.Sprintf(
		"cannot store non-storable type: `%s`",
		e.Type.QualifiedString(),
	)
}

// InterfaceMissingLocation is reported during interface lookup,
// if an interface is looked up without a location
type InterfaceMissingLocationError struct {
	QualifiedIdentifier string
}

var _ errors.UserError = InterfaceMissingLocationError{}

func (InterfaceMissingLocationError) IsUserError() {}

func (e InterfaceMissingLocationError) Error() string {
	return fmt.Sprintf(
		"tried to look up interface %s without a location",
		e.QualifiedIdentifier,
	)
}

// InvalidOperandsError
type InvalidOperandsError struct {
	LocationRange
	LeftType     StaticType
	RightType    StaticType
	FunctionName string
	Operation    ast.Operation
}

var _ errors.UserError = InvalidOperandsError{}

func (InvalidOperandsError) IsUserError() {}

func (e InvalidOperandsError) Error() string {
	var op string
	if e.Operation == ast.OperationUnknown {
		op = e.FunctionName
	} else {
		op = e.Operation.Symbol()
	}

	return fmt.Sprintf(
		"cannot apply operation %s to types: `%s`, `%s`",
		op,
		e.LeftType.String(),
		e.RightType.String(),
	)
}

// InvalidPublicKeyError is reported during PublicKey creation, if the PublicKey is invalid.
type InvalidPublicKeyError struct {
	PublicKey *ArrayValue
	Err       error
	LocationRange
}

var _ errors.UserError = InvalidPublicKeyError{}

func (InvalidPublicKeyError) IsUserError() {}

func (e InvalidPublicKeyError) Error() string {
	return fmt.Sprintf("invalid public key: %s, err: %s", e.PublicKey, e.Err)
}

func (e InvalidPublicKeyError) Unwrap() error {
	return e.Err
}

// NonTransferableValueError
type NonTransferableValueError struct {
	Value Value
}

var _ errors.UserError = NonTransferableValueError{}

func (NonTransferableValueError) IsUserError() {}

func (e NonTransferableValueError) Error() string {
	return "cannot transfer non-transferable value"
}

// DuplicateKeyInResourceDictionaryError
type DuplicateKeyInResourceDictionaryError struct {
	LocationRange
}

var _ errors.UserError = DuplicateKeyInResourceDictionaryError{}

func (DuplicateKeyInResourceDictionaryError) IsUserError() {}

func (e DuplicateKeyInResourceDictionaryError) Error() string {
	return "duplicate key in resource dictionary"
}

// StorageMutatedDuringIterationError
type StorageMutatedDuringIterationError struct {
	LocationRange
}

var _ errors.UserError = StorageMutatedDuringIterationError{}

func (StorageMutatedDuringIterationError) IsUserError() {}

func (StorageMutatedDuringIterationError) Error() string {
	return "storage iteration continued after modifying storage"
}

// ContainerMutatedDuringIterationError
type ContainerMutatedDuringIterationError struct {
	LocationRange
}

var _ errors.UserError = ContainerMutatedDuringIterationError{}

func (ContainerMutatedDuringIterationError) IsUserError() {}

func (ContainerMutatedDuringIterationError) Error() string {
	return "resource container modified during iteration"
}

// InvalidHexByteError
type InvalidHexByteError struct {
	LocationRange
	Byte byte
}

var _ errors.UserError = InvalidHexByteError{}

func (InvalidHexByteError) IsUserError() {}

func (e InvalidHexByteError) Error() string {
	return fmt.Sprintf("invalid byte in hex string: %x", e.Byte)
}

// InvalidHexLengthError
type InvalidHexLengthError struct {
	LocationRange
}

var _ errors.UserError = InvalidHexLengthError{}

func (InvalidHexLengthError) IsUserError() {}

func (InvalidHexLengthError) Error() string {
	return "hex string has non-even length"
}

// InvalidatedResourceReferenceError is reported when accessing a reference value
// that is pointing to a moved or destroyed resource.
type InvalidatedResourceReferenceError struct {
	LocationRange
}

var _ errors.UserError = InvalidatedResourceReferenceError{}

func (InvalidatedResourceReferenceError) IsUserError() {}

func (e InvalidatedResourceReferenceError) Error() string {
	return "referenced resource has been moved or destroyed after taking the reference"
}

// DuplicateAttachmentError
type DuplicateAttachmentError struct {
	AttachmentType sema.Type
	Value          *CompositeValue
	LocationRange
}

var _ errors.UserError = DuplicateAttachmentError{}

func (DuplicateAttachmentError) IsUserError() {}

func (e DuplicateAttachmentError) Error() string {
	return fmt.Sprintf(
		"cannot attach %s to %s, as it already exists on that value",
		e.AttachmentType.QualifiedString(),
		e.Value.QualifiedIdentifier,
	)
}

// AttachmentIterationMutationError
type AttachmentIterationMutationError struct {
	Value *CompositeValue
	LocationRange
}

var _ errors.UserError = AttachmentIterationMutationError{}

func (AttachmentIterationMutationError) IsUserError() {}

func (e AttachmentIterationMutationError) Error() string {
	return fmt.Sprintf(
		"cannot modify %s's attachments while iterating over them",
		e.Value.QualifiedIdentifier,
	)
}

// InvalidAttachmentOperationTargetError
type InvalidAttachmentOperationTargetError struct {
	Value Value
	LocationRange
}

var _ errors.InternalError = InvalidAttachmentOperationTargetError{}

func (InvalidAttachmentOperationTargetError) IsInternalError() {}

func (e InvalidAttachmentOperationTargetError) Error() string {
	return fmt.Sprintf(
		"%s cannot add or remove attachment with non-owned value (%T)",
		errors.InternalErrorMessagePrefix,
		e.Value,
	)
}

// RecursiveTransferError
type RecursiveTransferError struct {
	LocationRange
}

var _ errors.UserError = RecursiveTransferError{}

func (RecursiveTransferError) IsUserError() {}

func (RecursiveTransferError) Error() string {
	return "recursive transfer of value"
}

func WrappedExternalError(err error) error {
	switch err := err.(type) {
	case
		// If the error is a go-runtime error, don't wrap.
		// These are crashers.
		runtime.Error,

		// If the error is already a cadence error, then avoid redundant wrapping.
		errors.InternalError,
		errors.UserError,
		errors.ExternalError,
		Error:
		return err

	default:
		return errors.NewExternalError(err)
	}
}

// CapabilityAddressPublishingError
type CapabilityAddressPublishingError struct {
	LocationRange
	CapabilityAddress AddressValue
	AccountAddress    AddressValue
}

var _ errors.UserError = CapabilityAddressPublishingError{}

func (CapabilityAddressPublishingError) IsUserError() {}

func (e CapabilityAddressPublishingError) Error() string {
	return fmt.Sprintf(
		"cannot publish capability of account %s in account %s",
		e.CapabilityAddress.String(),
		e.AccountAddress.String(),
	)
}

// EntitledCapabilityPublishingError
type EntitledCapabilityPublishingError struct {
	LocationRange
	BorrowType *ReferenceStaticType
	Path       PathValue
}

var _ errors.UserError = EntitledCapabilityPublishingError{}

func (EntitledCapabilityPublishingError) IsUserError() {}

func (e EntitledCapabilityPublishingError) Error() string {
	return fmt.Sprintf(
		"cannot publish capability of type `%s` to the path %s",
		e.BorrowType.ID(),
		e.Path.String(),
	)
}

// NestedReferenceError
type NestedReferenceError struct {
	Value ReferenceValue
	LocationRange
}

var _ errors.UserError = NestedReferenceError{}

func (NestedReferenceError) IsUserError() {}

func (e NestedReferenceError) Error() string {
	return fmt.Sprintf(
		"cannot create a nested reference to %s",
		e.Value.String(),
	)
}

// InclusiveRangeConstructionError

type InclusiveRangeConstructionError struct {
	LocationRange
	Message string
}

var _ errors.UserError = InclusiveRangeConstructionError{}

func (InclusiveRangeConstructionError) IsUserError() {}

func (e InclusiveRangeConstructionError) Error() string {
	const message = "InclusiveRange construction failed"
	if e.Message == "" {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Message)
}

// InvalidCapabilityIssueTypeError
type InvalidCapabilityIssueTypeError struct {
	ExpectedTypeDescription string
	ActualType              sema.Type
	LocationRange
}

var _ errors.UserError = InvalidCapabilityIssueTypeError{}

func (InvalidCapabilityIssueTypeError) IsUserError() {}

func (e InvalidCapabilityIssueTypeError) Error() string {
	return fmt.Sprintf(
		"invalid type: expected %s, got `%s`",
		e.ExpectedTypeDescription,
		e.ActualType.QualifiedString(),
	)
}

// ResourceReferenceDereferenceError
type ResourceReferenceDereferenceError struct {
	LocationRange
}

var _ errors.InternalError = ResourceReferenceDereferenceError{}

func (ResourceReferenceDereferenceError) IsInternalError() {}

func (e ResourceReferenceDereferenceError) Error() string {
	return fmt.Sprintf(
		"%s resource-references cannot be dereferenced",
		errors.InternalErrorMessagePrefix,
	)
}

// ResourceLossError
type ResourceLossError struct {
	LocationRange
}

var _ errors.UserError = ResourceLossError{}

func (ResourceLossError) IsUserError() {}

func (e ResourceLossError) Error() string {
	return "resource loss: attempting to assign to non-nil resource-typed value"
}

// InvalidCapabilityIDError

type InvalidCapabilityIDError struct{}

var _ errors.InternalError = InvalidCapabilityIDError{}

func (InvalidCapabilityIDError) IsInternalError() {}

func (e InvalidCapabilityIDError) Error() string {
	return fmt.Sprintf(
		"%s capability created with invalid ID",
		errors.InternalErrorMessagePrefix,
	)
}

// ReferencedValueChangedError
type ReferencedValueChangedError struct {
	LocationRange
}

var _ errors.UserError = ReferencedValueChangedError{}

func (ReferencedValueChangedError) IsUserError() {}

func (e ReferencedValueChangedError) Error() string {
	return "referenced value has been changed after taking the reference"
}

// GetCapabilityError
type GetCapabilityError struct {
	LocationRange
}

var _ errors.UserError = GetCapabilityError{}

func (GetCapabilityError) IsUserError() {}

func (e GetCapabilityError) Error() string {
	return "cannot get capability"
}
