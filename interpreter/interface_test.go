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

package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence/activations"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/sema"
	"github.com/onflow/cadence/stdlib"
	. "github.com/onflow/cadence/test_utils/common_utils"
)

func TestInterpretInterfaceDefaultImplementation(t *testing.T) {

	t.Parallel()

	t.Run("interface", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `

          struct interface IA {
              fun test(): Int {
                  return 42
              }
          }

          struct Test: IA {

          }

          fun main(): Int {
              return Test().test()
          }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(42),
			value,
		)
	})

	t.Run("interface variable", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface IA {
                let x: Int
                fun getX(): Int {
                    return self.x
                }
            }

            struct Foo: IA {
                let x: Int
                init() {
                    self.x = 123
                }
           }

            struct Bar: IA {
                let x: Int
                init() {
                    self.x = 456
                }
            }

            fun test(): [Int;2] {
                let foo = Foo()
                let bar = Bar()

                return [foo.getX(), bar.getX()]
            }
        `)

		value, err := inter.Invoke("test")
		require.NoError(t, err)

		require.IsType(t, &interpreter.ArrayValue{}, value)
		array := value.(*interpreter.ArrayValue)

		// Check here whether:
		//  - The value set for `x` by the implementation is correctly set/returned.
		//  - Correct variable scope is used / Scopes are not shared.
		//    i.e: Value set by `Foo` doesn't affect `Bar`, and vice-versa

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(123),
			array.Get(inter, interpreter.EmptyLocationRange, 0),
		)
		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(456),
			array.Get(inter, interpreter.EmptyLocationRange, 1),
		)
	})

	t.Run("inherited interface function", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `

          struct interface I {
              fun test(): Int {
				return 3
			  }
          }

          struct interface J: I {}

		  struct S: J {}

		  fun foo(_ s: {J}): Int {
			return s.test()
		  }

          fun main(): Int {
			return foo(S())
          }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})

	t.Run("interface method subtyping", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
          struct A {
              var bar: Int
              init() {
                  self.bar = 4
              }
          }

          struct interface I {
              fun foo(): A?
          }

          struct S: I {
              fun foo(): A {
                  return A()
              }
          }

          fun main(): Int? {
              var s: {I} = S()
              return s.foo()?.bar
          }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredSomeValueNonCopying(
				interpreter.NewUnmeteredIntValueFromInt64(4),
			),
			value,
		)
	})
}

func TestInterpretInterfaceDefaultImplementationWhenOverridden(t *testing.T) {

	t.Parallel()

	t.Run("interface", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `

          struct interface IA {
              fun test(): Int {
                  return 41
              }
          }

          struct Test: IA {
              fun test(): Int {
                  return 42
              }
          }

          fun main(): Int {
              return Test().test()
          }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(42),
			value,
		)
	})

}

func TestInterpretInterfaceInheritance(t *testing.T) {

	t.Parallel()

	t.Run("struct interface", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                let x: Int

                fun test(): Int
            }

            struct interface B: A {}

            struct C: B {
                let x: Int

                init() {
                    self.x = 3
                }

                fun test(): Int {
                    return self.x
                }
            }

            access(all) fun main(): Int {
                let c = C()
                return c.test()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})

	t.Run("resource interface", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            resource interface A {
                let x: Int

                fun test(): Int
            }

            resource interface B: A {}

            resource C: B {
                let x: Int

                init() {
                    self.x = 3
                }

                fun test(): Int {
                    return self.x
                }
            }

            access(all) fun main(): Int {
                let c <- create C()
                let x = c.test()
                destroy c
                return x
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})

	t.Run("duplicate methods", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(): Int
            }

            struct interface B: A {
                access(all) fun test(): Int
            }

            struct C: B {
                fun test(): Int {
                    return 3
                }
            }

            access(all) fun main(): Int {
                let c = C()
                return c.test()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})

	t.Run("indirect default method", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(): Int {
                    return 3
                }
            }

            struct interface B: A {}

            struct C: B {}

            access(all) fun main(): Int {
                let c = C()
                return c.test()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})

	t.Run("default method via different paths", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(): Int {
                    return 3
                }
            }

            struct interface B: A {}

            struct interface C: A {}

            struct D: B, C {}

            access(all) fun main(): Int {
                let d = D()
                return d.test()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(3),
			value,
		)
	})
}

func TestInterpretInterfaceFunctionConditionsInheritance(t *testing.T) {

	t.Parallel()

	t.Run("condition in super", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(_ a: Int): Int {
                    pre { a > 10 }
                }
            }

            struct interface B: A {
                access(all) fun test(_ a: Int): Int
            }

            struct C: B {
                fun test(_ a: Int): Int {
                    return a + 3
                }
            }

            access(all) fun main(_ a: Int): Int {
                let c = C()
                return c.test(a)
            }
        `)

		value, err := inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(15))
		require.NoError(t, err)
		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(18),
			value,
		)

		// Implementation should satisfy inherited conditions
		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(5))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})
	})

	t.Run("condition in child", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(_ a: Int): Int
            }

            struct interface B: A {
                access(all) fun test(_ a: Int): Int {
                    pre { a > 10 }
                }
            }

            struct C: B {
                fun test(_ a: Int): Int {
                    return a + 3
                }
            }

            access(all) fun main(_ a: Int): Int {
                let c = C()
                return c.test(a)
            }
        `)

		value, err := inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(15))
		require.NoError(t, err)
		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(18),
			value,
		)

		// Implementation should satisfy inherited conditions
		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(5))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})
	})

	t.Run("conditions in both", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(_ a: Int): Int {
                    pre { a < 20 }
                }
            }

            struct interface B: A {
                access(all) fun test(_ a: Int): Int {
                    pre { a > 10 }
                }
            }

            struct C: B {
                fun test(_ a: Int): Int {
                    return a + 3
                }
            }

            access(all) fun main(_ a: Int): Int {
                let c = C()
                return c.test(a)
            }
        `)

		value, err := inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(15))
		require.NoError(t, err)
		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(18),
			value,
		)

		// Implementation should satisfy both inherited conditions

		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(5))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})

		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(25))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})
	})

	t.Run("conditions from two paths", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            struct interface A {
                access(all) fun test(_ a: Int): Int {
                    pre { a < 20 }
                }
            }

            struct interface B {
                access(all) fun test(_ a: Int): Int {
                    pre { a > 10 }
                }
            }

            struct interface C: A, B {}

            struct D: C {
                fun test(_ a: Int): Int {
                    return a + 3
                }
            }

            access(all) fun main(_ a: Int): Int {
                let d = D()
                return d.test(a)
            }
        `)

		value, err := inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(15))
		require.NoError(t, err)
		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(18),
			value,
		)

		// Implementation should satisfy both inherited conditions

		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(5))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})

		_, err = inter.Invoke("main", interpreter.NewUnmeteredIntValueFromInt64(25))
		RequireError(t, err)
		assert.ErrorAs(t, err, &interpreter.ConditionError{})
	})

	t.Run("pre conditions order", func(t *testing.T) {

		t.Parallel()

		logFunctionType := sema.NewSimpleFunctionType(
			sema.FunctionPurityView,
			[]sema.Parameter{
				{
					Label:          sema.ArgumentLabelNotRequired,
					Identifier:     "value",
					TypeAnnotation: sema.AnyStructTypeAnnotation,
				},
			},
			sema.VoidTypeAnnotation,
		)

		var logs []string
		valueDeclaration := stdlib.NewStandardLibraryStaticFunction(
			"log",
			logFunctionType,
			"",
			func(invocation interpreter.Invocation) interpreter.Value {
				msg := invocation.Arguments[0].(*interpreter.StringValue).Str
				logs = append(logs, msg)
				return interpreter.Void
			},
		)

		baseValueActivation := sema.NewVariableActivation(sema.BaseValueActivation)
		baseValueActivation.DeclareValue(valueDeclaration)
		baseActivation := activations.NewActivation(nil, interpreter.BaseActivation)
		interpreter.Declare(baseActivation, valueDeclaration)

		// Inheritance hierarchy is as follows:
		//
		//       A (concrete type)
		//       |
		//       B (interface)
		//      / \
		//     C   D
		//    / \ /
		//   E   F

		inter, err := parseCheckAndInterpretWithOptions(t, `
            struct A: B {
                access(all) fun test() {
                    pre { print("A") }
                }
            }

            struct interface B: C, D {
                access(all) fun test() {
                    pre { print("B") }
                }
            }

            struct interface C: E, F {
                access(all) fun test() {
                    pre { print("C") }
                }
            }

            struct interface D: F {
                access(all) fun test() {
                    pre { print("D") }
                }
            }

            struct interface E {
                access(all) fun test() {
                    pre { print("E") }
                }
            }

            struct interface F {
                access(all) fun test() {
                    pre { print("F") }
                }
            }

            access(all) view fun print(_ msg: String): Bool {
                log(msg)
                return true
            }

            access(all) fun main() {
                let a = A()
                a.test()
            }`,
			ParseCheckAndInterpretOptions{
				CheckerConfig: &sema.Config{
					BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
						return baseValueActivation
					},
				},
				Config: &interpreter.Config{
					BaseActivationHandler: func(common.Location) *interpreter.VariableActivation {
						return baseActivation
					},
				},
			},
		)
		require.NoError(t, err)

		_, err = inter.Invoke("main")
		require.NoError(t, err)

		// The pre-conditions of the interfaces are executed first, with depth-first pre-order traversal.
		// The pre-condition of the concrete type is executed at the end, after the interfaces.
		assert.Equal(t, []string{"B", "C", "E", "F", "D", "A"}, logs)
	})

	t.Run("post conditions order", func(t *testing.T) {

		t.Parallel()

		logFunctionType := sema.NewSimpleFunctionType(
			sema.FunctionPurityView,
			[]sema.Parameter{
				{
					Label:          sema.ArgumentLabelNotRequired,
					Identifier:     "value",
					TypeAnnotation: sema.AnyStructTypeAnnotation,
				},
			},
			sema.VoidTypeAnnotation,
		)

		var logs []string
		valueDeclaration := stdlib.NewStandardLibraryStaticFunction(
			"log",
			logFunctionType,
			"",
			func(invocation interpreter.Invocation) interpreter.Value {
				msg := invocation.Arguments[0].(*interpreter.StringValue).Str
				logs = append(logs, msg)
				return interpreter.Void
			},
		)

		baseValueActivation := sema.NewVariableActivation(sema.BaseValueActivation)
		baseValueActivation.DeclareValue(valueDeclaration)
		baseActivation := activations.NewActivation(nil, interpreter.BaseActivation)
		interpreter.Declare(baseActivation, valueDeclaration)

		// Inheritance hierarchy is as follows:
		//
		//       A (concrete type)
		//       |
		//       B (interface)
		//      / \
		//     C   D
		//    / \ /
		//   E   F

		inter, err := parseCheckAndInterpretWithOptions(t, `
            struct A: B {
                access(all) fun test() {
                    post { print("A") }
                }
            }

            struct interface B: C, D {
                access(all) fun test() {
                    post { print("B") }
                }
            }

            struct interface C: E, F {
                access(all) fun test() {
                    post { print("C") }
                }
            }

            struct interface D: F {
                access(all) fun test() {
                    post { print("D") }
                }
            }

            struct interface E {
                access(all) fun test() {
                    post { print("E") }
                }
            }

            struct interface F {
                access(all) fun test() {
                    post { print("F") }
                }
            }

            access(all) view fun print(_ msg: String): Bool {
                log(msg)
                return true
            }

            access(all) fun main() {
                let a = A()
                a.test()
            }`,
			ParseCheckAndInterpretOptions{
				CheckerConfig: &sema.Config{
					BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
						return baseValueActivation
					},
				},
				Config: &interpreter.Config{
					BaseActivationHandler: func(common.Location) *interpreter.VariableActivation {
						return baseActivation
					},
				},
			},
		)
		require.NoError(t, err)

		_, err = inter.Invoke("main")
		require.NoError(t, err)

		// The post-condition of the concrete type is executed first, before the interfaces.
		// The post-conditions of the interfaces are executed after that, with the reversed depth-first pre-order.
		assert.Equal(t, []string{"A", "D", "F", "E", "C", "B"}, logs)
	})

	t.Run("nested resource interface unrelated", func(t *testing.T) {

		t.Parallel()

		logFunctionType := sema.NewSimpleFunctionType(
			sema.FunctionPurityView,
			[]sema.Parameter{
				{
					Label:          sema.ArgumentLabelNotRequired,
					Identifier:     "value",
					TypeAnnotation: sema.AnyStructTypeAnnotation,
				},
			},
			sema.VoidTypeAnnotation,
		)

		var logs []string
		valueDeclaration := stdlib.NewStandardLibraryStaticFunction(
			"log",
			logFunctionType,
			"",
			func(invocation interpreter.Invocation) interpreter.Value {
				msg := invocation.Arguments[0].(*interpreter.StringValue).Str
				logs = append(logs, msg)
				return interpreter.Void
			},
		)

		baseValueActivation := sema.NewVariableActivation(sema.BaseValueActivation)
		baseValueActivation.DeclareValue(valueDeclaration)
		baseActivation := activations.NewActivation(nil, interpreter.BaseActivation)
		interpreter.Declare(baseActivation, valueDeclaration)

		inter, err := parseCheckAndInterpretWithOptions(t, `
            contract interface A {
                struct interface Nested {
                    access(all) fun test(): Int {
                        post { print("A") }
                    }
                }
            }

            contract interface B: A {
                struct interface Nested {
                    access(all) fun test(): String {
                        post { print("B") }
                    }
                }
            }
            
            contract C {
                struct Nested: B.Nested {
                    fun test(): String {
                        return "C"
                    }
                }
            }

            access(all) view fun print(_ msg: String): Bool {
                log(msg)
                return true
            }

            access(all) fun main() {
                let n = C.Nested()
                n.test()
            }
        `,
			ParseCheckAndInterpretOptions{
				CheckerConfig: &sema.Config{
					BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
						return baseValueActivation
					},
				},
				Config: &interpreter.Config{
					BaseActivationHandler: func(common.Location) *interpreter.VariableActivation {
						return baseActivation
					},
					ContractValueHandler: makeContractValueHandler(nil, nil, nil),
				},
			},
		)
		require.NoError(t, err)

		_, err = inter.Invoke("main")
		require.NoError(t, err)

		// A.Nested and B.Nested are two distinct separate functions
		assert.Equal(t, []string{"B"}, logs)
	})

	t.Run("pre condition in parent, default impl in child", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            access(all) resource interface A {
                access(all) fun get(): Int {
                   pre {
                       true
                   }
                }
            }

            access(all) resource interface B: A {
                access(all) fun get(): Int {
                    return 4
                }
            }

            access(all) resource R: B {}

            access(all) fun main(): Int {
                let r <- create R()
                let value = r.get()
                destroy r
                return value
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(4),
			value,
		)
	})

	t.Run("post condition in parent, default impl in child", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            access(all) resource interface A {
                access(all) fun get(): Int {
                   post {
                       true
                   }
                }
            }

            access(all) resource interface B: A {
                access(all) fun get(): Int {
                    return 4
                }
            }

            access(all) resource R: B {}

            access(all) fun main(): Int {
                let r <- create R()
                let value = r.get()
                destroy r
                return value
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(4),
			value,
		)
	})

	t.Run("siblings with condition in first and default impl in second", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            access(all) struct interface A {
                access(all) fun get(): Int {
                   post { true }
                }
            }

            access(all) struct interface B {
                access(all) fun get(): Int {
                    return 4
                }
            }

            struct interface C: A, B {}

            access(all) struct S: C {}

            access(all) fun main(): Int {
                let s = S()
                return s.get()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(4),
			value,
		)
	})

	t.Run("siblings with default impl in first and condition in second", func(t *testing.T) {

		t.Parallel()

		inter := parseCheckAndInterpret(t, `
            access(all) struct interface A {
                access(all) fun get(): Int {
                    return 4
                }
            }

            access(all) struct interface B {
                access(all) fun get(): Int {
                   post { true }
                }
            }

            struct interface C: A, B {}

            access(all) struct S: C {}

            access(all) fun main(): Int {
                let s = S()
                return s.get()
            }
        `)

		value, err := inter.Invoke("main")
		require.NoError(t, err)

		assert.Equal(t,
			interpreter.NewUnmeteredIntValueFromInt64(4),
			value,
		)
	})

	t.Run("result variable in conditions", func(t *testing.T) {

		t.Parallel()

		inter, getLogs, err := parseCheckAndInterpretWithLogs(t, `
            access(all) resource interface I1 {
                access(all) let s: String

                access(all) fun echo(_ s: String): String {
                    post {
                        result == self.s: "result must match stored input, got: ".concat(result)
                    }
                }
            }

            access(all) resource interface I2: I1 {
                access(all) let s: String

                access(all) fun echo(_ s: String): String {
                    log(s)
                    return self.s
                }
            }

            access(all) resource R: I2 {
                access(all) let s: String

                init() {
                    self.s = "hello"
                }
            }

            access(all) fun main() {
                let r <- create R()
                r.echo("hello")
                destroy r
            }
        `)
		require.NoError(t, err)

		_, err = inter.Invoke("main")
		require.NoError(t, err)

		logs := getLogs()
		require.Len(t, logs, 1)
		assert.Equal(t, "\"hello\"", logs[0])
	})

	t.Run("default and conditions in parent, more conditions in child", func(t *testing.T) {

		t.Parallel()

		var logs []string

		logFunction := stdlib.NewStandardLibraryStaticFunction(
			"log",
			&sema.FunctionType{
				Parameters: []sema.Parameter{
					{
						Label:          sema.ArgumentLabelNotRequired,
						Identifier:     "value",
						TypeAnnotation: sema.NewTypeAnnotation(sema.AnyStructType),
					},
				},
				ReturnTypeAnnotation: sema.NewTypeAnnotation(
					sema.VoidType,
				),
				Purity: sema.FunctionPurityView,
			},
			``,
			func(invocation interpreter.Invocation) interpreter.Value {
				message := invocation.Arguments[0].MeteredString(
					invocation.Interpreter,
					interpreter.SeenReferences{},
					invocation.LocationRange,
				)
				logs = append(logs, message)
				return interpreter.Void
			},
		)

		baseValueActivation := sema.NewVariableActivation(sema.BaseValueActivation)
		baseValueActivation.DeclareValue(logFunction)

		baseActivation := activations.NewActivation(nil, interpreter.BaseActivation)
		interpreter.Declare(baseActivation, logFunction)

		code := `
            struct interface Foo {
                fun test() {
                    pre {
                         printMessage("invoked Foo.test() pre-condition")
                    }
                    post {
                         printMessage("invoked Foo.test() post-condition")
                    }
                    printMessage("invoked Foo.test()")
                }
            }

            struct Test: Foo {
            }

            access(all) view fun printMessage(_ msg: String): Bool {
                log(msg)
                return true
            }

            fun main() {
               Test().test()
            }
        `

		inter, err := parseCheckAndInterpretWithOptions(
			t,
			code,
			ParseCheckAndInterpretOptions{
				Config: &interpreter.Config{
					BaseActivationHandler: func(_ common.Location) *interpreter.VariableActivation {
						return baseActivation
					},
				},
				CheckerConfig: &sema.Config{
					BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
						return baseValueActivation
					},
				},
				HandleCheckerError: nil,
			},
		)
		require.NoError(t, err)

		_, err = inter.Invoke("main")
		require.NoError(t, err)

		require.Equal(
			t,
			[]string{
				"\"invoked Foo.test() pre-condition\"",
				"\"invoked Foo.test()\"",
				"\"invoked Foo.test() post-condition\"",
			}, logs,
		)
	})
}

func TestInterpretNestedInterfaceCast(t *testing.T) {

	t.Parallel()

	inter, err := parseCheckAndInterpretWithOptions(t, `
	access(all) contract C {
		access(all) resource interface TopInterface {}
		access(all) resource interface MiddleInterface: TopInterface {}
		access(all) resource ConcreteResource: MiddleInterface {}
	 
		access(all) fun createMiddleInterface(): @{MiddleInterface} {
			return <-create ConcreteResource()
		}
	 }

	 access(all) fun main() {
		let x <- C.createMiddleInterface()
		let y <- x as! @{C.TopInterface}
		destroy y
	 }
        `,
		ParseCheckAndInterpretOptions{
			CheckerConfig: &sema.Config{},
			Config: &interpreter.Config{
				ContractValueHandler: makeContractValueHandler(nil, nil, nil),
			},
		},
	)
	require.NoError(t, err)

	_, err = inter.Invoke("main")
	require.NoError(t, err)
}
