// Code generated from storage_capability_controller.cdc. DO NOT EDIT.
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

package sema

import "github.com/onflow/cadence/ast"

const StorageCapabilityControllerTypeCapabilityFieldName = "capability"

var StorageCapabilityControllerTypeCapabilityFieldType = &CapabilityType{}

const StorageCapabilityControllerTypeCapabilityFieldDocString = `
The capability that is controlled by this controller.
`

const StorageCapabilityControllerTypeTagFieldName = "tag"

var StorageCapabilityControllerTypeTagFieldType = StringType

const StorageCapabilityControllerTypeTagFieldDocString = `
An arbitrary "tag" for the controller.
For example, it could be used to describe the purpose of the capability.
Empty by default.
`

const StorageCapabilityControllerTypeSetTagFunctionName = "setTag"

var StorageCapabilityControllerTypeSetTagFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "tag",
			TypeAnnotation: NewTypeAnnotation(StringType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		VoidType,
	),
}

const StorageCapabilityControllerTypeSetTagFunctionDocString = `
Updates this controller's tag to the provided string
`

const StorageCapabilityControllerTypeBorrowTypeFieldName = "borrowType"

var StorageCapabilityControllerTypeBorrowTypeFieldType = MetaType

const StorageCapabilityControllerTypeBorrowTypeFieldDocString = `
The type of the controlled capability, i.e. the T in ` + "`Capability<T>`" + `.
`

const StorageCapabilityControllerTypeCapabilityIDFieldName = "capabilityID"

var StorageCapabilityControllerTypeCapabilityIDFieldType = UInt64Type

const StorageCapabilityControllerTypeCapabilityIDFieldDocString = `
The identifier of the controlled capability.
All copies of a capability have the same ID.
`

const StorageCapabilityControllerTypeDeleteFunctionName = "delete"

var StorageCapabilityControllerTypeDeleteFunctionType = &FunctionType{
	ReturnTypeAnnotation: NewTypeAnnotation(
		VoidType,
	),
}

const StorageCapabilityControllerTypeDeleteFunctionDocString = `
Delete this capability controller,
and disable the controlled capability and its copies.

The controller will be deleted from storage,
but the controlled capability and its copies remain.

Once this function returns, the controller is no longer usable,
all further operations on the controller will panic.

Borrowing from the controlled capability or its copies will return nil.
`

const StorageCapabilityControllerTypeTargetFunctionName = "target"

var StorageCapabilityControllerTypeTargetFunctionType = &FunctionType{
	ReturnTypeAnnotation: NewTypeAnnotation(
		StoragePathType,
	),
}

const StorageCapabilityControllerTypeTargetFunctionDocString = `
Returns the targeted storage path of the controlled capability.
`

const StorageCapabilityControllerTypeRetargetFunctionName = "retarget"

var StorageCapabilityControllerTypeRetargetFunctionType = &FunctionType{
	Parameters: []Parameter{
		{
			Label:          ArgumentLabelNotRequired,
			Identifier:     "target",
			TypeAnnotation: NewTypeAnnotation(StoragePathType),
		},
	},
	ReturnTypeAnnotation: NewTypeAnnotation(
		VoidType,
	),
}

const StorageCapabilityControllerTypeRetargetFunctionDocString = `
Retarget the controlled capability to the given storage path.
The path may be different or the same as the current path.
`

const StorageCapabilityControllerTypeName = "StorageCapabilityController"

var StorageCapabilityControllerType = &SimpleType{
	Name:          StorageCapabilityControllerTypeName,
	QualifiedName: StorageCapabilityControllerTypeName,
	TypeID:        StorageCapabilityControllerTypeName,
	TypeTag:       StorageCapabilityControllerTypeTag,
	IsResource:    false,
	Storable:      false,
	Primitive:     false,
	Equatable:     false,
	Comparable:    false,
	Exportable:    false,
	Importable:    false,
	ContainFields: true,
}

func init() {
	StorageCapabilityControllerType.Members = func(t *SimpleType) map[string]MemberResolver {
		return MembersAsResolvers([]*Member{
			NewUnmeteredFieldMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				ast.VariableKindConstant,
				StorageCapabilityControllerTypeCapabilityFieldName,
				StorageCapabilityControllerTypeCapabilityFieldType,
				StorageCapabilityControllerTypeCapabilityFieldDocString,
			),
			NewUnmeteredFieldMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				ast.VariableKindVariable,
				StorageCapabilityControllerTypeTagFieldName,
				StorageCapabilityControllerTypeTagFieldType,
				StorageCapabilityControllerTypeTagFieldDocString,
			),
			NewUnmeteredFunctionMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				StorageCapabilityControllerTypeSetTagFunctionName,
				StorageCapabilityControllerTypeSetTagFunctionType,
				StorageCapabilityControllerTypeSetTagFunctionDocString,
			),
			NewUnmeteredFieldMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				ast.VariableKindConstant,
				StorageCapabilityControllerTypeBorrowTypeFieldName,
				StorageCapabilityControllerTypeBorrowTypeFieldType,
				StorageCapabilityControllerTypeBorrowTypeFieldDocString,
			),
			NewUnmeteredFieldMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				ast.VariableKindConstant,
				StorageCapabilityControllerTypeCapabilityIDFieldName,
				StorageCapabilityControllerTypeCapabilityIDFieldType,
				StorageCapabilityControllerTypeCapabilityIDFieldDocString,
			),
			NewUnmeteredFunctionMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				StorageCapabilityControllerTypeDeleteFunctionName,
				StorageCapabilityControllerTypeDeleteFunctionType,
				StorageCapabilityControllerTypeDeleteFunctionDocString,
			),
			NewUnmeteredFunctionMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				StorageCapabilityControllerTypeTargetFunctionName,
				StorageCapabilityControllerTypeTargetFunctionType,
				StorageCapabilityControllerTypeTargetFunctionDocString,
			),
			NewUnmeteredFunctionMember(
				t,
				PrimitiveAccess(ast.AccessAll),
				StorageCapabilityControllerTypeRetargetFunctionName,
				StorageCapabilityControllerTypeRetargetFunctionType,
				StorageCapabilityControllerTypeRetargetFunctionDocString,
			),
		})
	}
}
