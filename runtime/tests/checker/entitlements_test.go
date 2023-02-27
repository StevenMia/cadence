/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Dapper Labs, Inc.
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

package checker

import (
	"testing"

	"github.com/onflow/cadence/runtime/sema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckBasicEntitlementDeclaration(t *testing.T) {

	t.Parallel()

	t.Run("basic, no fields", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheck(t, `
			entitlement E {}
		`)

		assert.NoError(t, err)
		entitlement := checker.Elaboration.EntitlementType("S.test.E")
		assert.Equal(t, "E", entitlement.String())
		assert.Equal(t, 0, entitlement.Members.Len())
	})

	t.Run("basic, with fields", func(t *testing.T) {
		t.Parallel()
		checker, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
				var x: String
			}
		`)

		assert.NoError(t, err)
		entitlement := checker.Elaboration.EntitlementType("S.test.E")
		assert.Equal(t, "E", entitlement.String())
		assert.Equal(t, 2, entitlement.Members.Len())
	})

	t.Run("basic, with fun access modifier", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				pub fun foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementMemberAccessDeclaration{}, errs[0])
	})

	t.Run("basic, with field access modifier", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				access(self) let x: Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementMemberAccessDeclaration{}, errs[0])
	})

	t.Run("basic, with precondition", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo() {
					pre {

					}
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidEntitlementFunctionDeclaration{}, errs[0])
		require.IsType(t, &sema.InvalidImplementationError{}, errs[1])
	})

	t.Run("basic, with postcondition", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo() {
					post {

					}
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidEntitlementFunctionDeclaration{}, errs[0])
		require.IsType(t, &sema.InvalidImplementationError{}, errs[1])
	})

	t.Run("basic, with postconditions", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo() {
					post {
						1 == 2: "beep"
					}
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementFunctionDeclaration{}, errs[0])
	})

	t.Run("basic, with empty body", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidEntitlementFunctionDeclaration{}, errs[0])
		require.IsType(t, &sema.InvalidImplementationError{}, errs[1])

	})

	t.Run("basic, enum case", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				pub case green
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no nested resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				resource R {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no nested event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				event Foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no nested struct interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				struct interface R {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no nested entitlement", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				entitlement F {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no destroy", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				destroy()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("no special function", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				x()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementNestedDeclarationError{}, errs[0])
	})

	t.Run("priv access", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			priv entitlement E {
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})

	t.Run("duped members", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let x: Int
				fun x() 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.RedeclarationError{}, errs[0])
	})

	t.Run("invalid resource annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let x: @Int
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidResourceAnnotationError{}, errs[0])
	})

	t.Run("invalid destroy name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let destroy: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNameError{}, errs[0])
	})

	t.Run("invalid init name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let init: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNameError{}, errs[0])
	})
}

func TestCheckEntitlementDeclarationNesting(t *testing.T) {
	t.Parallel()
	t.Run("in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement E {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("in resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				entitlement E {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in resource interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource interface R {
				entitlement E {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct S {
				entitlement E {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				entitlement E {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
	})

	t.Run("in enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			enum X: UInt8 {
				entitlement E {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidNonEnumCaseError{}, errs[1])
	})
}

func TestCheckBasicEntitlementAccess(t *testing.T) {

	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let foo: String
			}
			struct interface S {
				access(E) let foo: String
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement A {
				let foo: String
			}
			entitlement B {
				let foo: String
				fun bar()
			}
			entitlement C {
				fun bar()
			}
			resource interface R {
				access(A, B) let foo: String
				access(B, C) fun bar()
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid in contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E {
					let foo: String
				}
				struct interface S {
					access(E) let foo: String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("valid in contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface C {
				entitlement E {
					let foo: String
				}
				struct interface S {
					access(E) let foo: String
				}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("qualified", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract C {
				entitlement E {
					let foo: String
					fun bar()
				}
				struct interface S {
					access(E) let foo: String
				}
			}
			resource R {
				access(C.E) fun bar() {}
			}
		`)

		assert.NoError(t, err)
	})
}

func TestCheckInvalidEntitlementAccess(t *testing.T) {

	t.Parallel()

	t.Run("invalid variable decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				var x: String
			}
			access(E) var x: String = ""
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid fun decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			access(E) fun foo() {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid contract field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			contract C {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid contract interface field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			contract interface C {
				access(E) fun foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource I {
				access(E) event Foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidNestedDeclarationError{}, errs[0])
		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[1])
	})

	t.Run("invalid enum case", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			enum X: UInt8 {
				access(E) case red
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidAccessModifierError{}, errs[0])
	})

	t.Run("missing entitlement declaration fun", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})

	t.Run("missing entitlement declaration field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				access(E) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.NotDeclaredError{}, errs[0])
	})
}

func TestCheckNonEntitlementAccess(t *testing.T) {

	t.Parallel()

	t.Run("resource", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("resource interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource interface E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("struct interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			resource E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("event", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			event E()
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("contract", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("contract interface", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			contract interface E {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})

	t.Run("enum", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			enum E: UInt8 {}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNonEntitlementAccessError{}, errs[0])
	})
}

func TestCheckEntitlementInheritance(t *testing.T) {

	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct S {
				access(E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("pub subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				pub fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("pub(set) subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				var x: String
			}
			struct interface I {
				pub(set) var x: String
			}
			struct S: I {
				access(E) var x: String
				init() {
					self.x = ""
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("pub supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				pub fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("pub(set) supertyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				var x: String
			}
			struct interface I {
				access(E) var x: String
			}
			struct S: I {
				pub(set) var x: String
				init() {
					self.x = ""
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access contract subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(contract) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access account subtyping invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(account) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access account supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(account) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("access contract supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				access(contract) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("priv supertying invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct S: I {
				priv fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("expanded entitlements valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E, F) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("expanded entitlements invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("expanded entitlements also invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(E, F) fun foo() 
			}
			struct S: I, J {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("different entitlements invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			entitlement G {
				fun foo()
			}
			struct interface I {
				access(E) fun foo() 
			}
			struct interface J {
				access(F) fun foo() 
			}
			struct S: I, J {
				access(E, G) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})

	t.Run("fewer entitlements invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			struct interface I {
				access(E, F) fun foo() 
			}
			struct S: I {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.ConformanceError{}, errs[0])
	})
}

func TestCheckEntitlementConformance(t *testing.T) {

	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			resource R {
				access(E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("unimplemented method", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
				fun bar()
			}
			resource R {
				access(E) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("unimplemented field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
				let x: String
			}
			resource interface R {
				access(E) fun foo() 
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("missing method", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
			}
			resource R {
				access(E) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementMemberNotDeclaredError{}, errs[0])
	})

	t.Run("missing field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
			}
			resource interface R {
				access(E) let x: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementMemberNotDeclaredError{}, errs[0])
	})

	t.Run("multiple entitlements", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo()
			}
			resource R {
				access(E, F) fun foo() {}
			}
		`)

		assert.NoError(t, err)
	})

	t.Run("multiple entitlements mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo()
			}
			entitlement F {
				fun foo(): String
			}
			resource R {
				access(E, F) fun foo() {}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
	})

	t.Run("multiple entitlements field mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let x: String
			}
			entitlement F {
				let x: UInt8
			}
			resource interface R {
				access(E, F) let x: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
	})

	t.Run("multiple entitlements mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let x: Bool
			}
			entitlement F {
				let x: UInt8
			}
			resource interface R {
				access(E, F) let x: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
		require.IsType(t, &sema.EntitlementConformanceError{}, errs[1])
	})

	t.Run("one missing one mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let x: Bool
			}
			entitlement F {
			}
			resource interface R {
				access(E, F) let x: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
		require.IsType(t, &sema.EntitlementMemberNotDeclaredError{}, errs[1])
	})

	t.Run("field does not match function", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				let foo: (fun ():Void)
			}
			resource interface R {
				access(E) fun foo()
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
	})

	t.Run("subtype invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {
				fun foo(): @{I}
			}
			resource interface I {
				access(E) fun foo(): @{I}
			}
			resource R {
				access(E) fun foo(): @R {
					return <-create R()
				}
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.EntitlementConformanceError{}, errs[0])
	})
}

func TestCheckEntitlementTypeAnnotation(t *testing.T) {

	t.Parallel()

	t.Run("invalid local annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			let x: E = ""
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("invalid param annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			pub fun foo(e: E) {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid return annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				pub fun foo(): E 
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid field annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				let e: E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid conformance annotation", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource R: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidConformanceError{}, errs[0])
	})

	t.Run("invalid array annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				let e: [E]
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid fun annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				let e: (fun (E): Void)
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("invalid enum conformance", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			enum X: E {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEnumRawTypeError{}, errs[0])
	})

	t.Run("invalid dict annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				let e: {E: E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		// key
		require.IsType(t, &sema.InvalidDictionaryKeyTypeError{}, errs[0])
		// value
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("invalid fun annot", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			resource interface I {
				let e: (fun (E): Void)
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("runtype type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			let e = Type<E>()
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("type arg", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E {}
			let e = authAccount.load<E>(from: /storage/foo)
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		// entitlements are not storable either
		require.IsType(t, &sema.TypeMismatchError{}, errs[1])
	})

	t.Run("restricted", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E {}
			resource interface I {
				let e: E{E}
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.InvalidRestrictionTypeError{}, errs[0])
		require.IsType(t, &sema.InvalidRestrictedTypeError{}, errs[1])
	})

	t.Run("reference", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E {}
			resource interface I {
				let e: &E
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})

	t.Run("capability", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E {}
			resource interface I {
				let e: Capability<&E>
			}
		`)

		errs := RequireCheckerErrors(t, err, 2)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[1])
	})

	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheckAccount(t, `
			entitlement E {}
			resource interface I {
				let e: E?
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.DirectEntitlementAnnotationError{}, errs[0])
	})
}
