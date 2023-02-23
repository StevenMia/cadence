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

func TestCheckInvalidEntitlementAccess(t *testing.T) {
	t.Run("invalid variable decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			access(E) var x: String = ""
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid fun decl", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
			access(E) fun foo() {}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidEntitlementAccessError{}, errs[0])
	})

	t.Run("invalid contract field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			entitlement E {}
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
			entitlement E {}
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
			contract interface I {
				access(E) event Foo()
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

		require.IsType(t, &sema.InvalidNameError{}, errs[0])
	})

	t.Run("missing entitlement declaration field", func(t *testing.T) {
		t.Parallel()
		_, err := ParseAndCheck(t, `
			struct interface S {
				access(E) let foo: String
			}
		`)

		errs := RequireCheckerErrors(t, err, 1)

		require.IsType(t, &sema.InvalidNameError{}, errs[0])
	})
}
