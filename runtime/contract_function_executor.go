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
	"sync"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/sema"
)

type interpreterContractFunctionExecutor struct {
	context          Context
	environment      Environment
	result           cadence.Value
	executeErr       error
	preprocessErr    error
	codesAndPrograms CodesAndPrograms
	runtime          *interpreterRuntime
	storage          *Storage
	contractLocation common.AddressLocation
	functionName     string
	arguments        []cadence.Value
	argumentTypes    []sema.Type
	executeOnce      sync.Once
	preprocessOnce   sync.Once
}

func newInterpreterContractFunctionExecutor(
	runtime *interpreterRuntime,
	contractLocation common.AddressLocation,
	functionName string,
	arguments []cadence.Value,
	argumentTypes []sema.Type,
	context Context,
) *interpreterContractFunctionExecutor {
	return &interpreterContractFunctionExecutor{
		runtime:          runtime,
		contractLocation: contractLocation,
		functionName:     functionName,
		arguments:        arguments,
		argumentTypes:    argumentTypes,
		context:          context,
	}
}

func (executor *interpreterContractFunctionExecutor) Preprocess() error {
	executor.preprocessOnce.Do(func() {
		executor.preprocessErr = executor.preprocess()
	})

	return executor.preprocessErr
}

func (executor *interpreterContractFunctionExecutor) Execute() error {
	executor.executeOnce.Do(func() {
		executor.result, executor.executeErr = executor.execute()
	})

	return executor.executeErr
}

func (executor *interpreterContractFunctionExecutor) Result() (cadence.Value, error) {
	// Note: Execute's error is saved into executor.executeErr and return in
	// the next line.
	_ = executor.Execute()
	return executor.result, executor.executeErr
}

func (executor *interpreterContractFunctionExecutor) preprocess() (err error) {
	context := executor.context
	location := context.Location

	codesAndPrograms := NewCodesAndPrograms()
	executor.codesAndPrograms = codesAndPrograms

	interpreterRuntime := executor.runtime

	defer interpreterRuntime.Recover(
		func(internalErr Error) {
			err = internalErr
		},
		location,
		codesAndPrograms,
	)

	runtimeInterface := context.Interface

	storage := NewStorage(
		runtimeInterface,
		runtimeInterface,
		StorageConfig{
			StorageFormatV2Enabled: interpreterRuntime.defaultConfig.StorageFormatV2Enabled,
		},
	)
	executor.storage = storage

	environment := context.Environment
	if environment == nil {
		environment = NewBaseInterpreterEnvironment(interpreterRuntime.defaultConfig)
	}
	environment.Configure(
		runtimeInterface,
		codesAndPrograms,
		storage,
		context.CoverageReport,
	)
	executor.environment = environment

	return nil
}

func (executor *interpreterContractFunctionExecutor) execute() (val cadence.Value, err error) {
	err = executor.Preprocess()
	if err != nil {
		return nil, err
	}

	environment := executor.environment
	context := executor.context
	location := context.Location
	codesAndPrograms := executor.codesAndPrograms
	interpreterRuntime := executor.runtime

	defer interpreterRuntime.Recover(
		func(internalErr Error) {
			err = internalErr
		},
		location,
		codesAndPrograms,
	)

	// create interpreter
	_, inter, err := environment.Interpret(
		location,
		nil,
		nil,
	)
	if err != nil {
		return nil, newError(err, location, codesAndPrograms)
	}

	// ensure the contract is loaded
	inter = inter.EnsureLoaded(executor.contractLocation)

	interpreterArguments := make([]interpreter.Value, len(executor.arguments))

	locationRange := interpreter.LocationRange{
		Location:    location,
		HasPosition: ast.EmptyRange,
	}

	for i, argumentType := range executor.argumentTypes {
		interpreterArguments[i], err = executor.convertArgument(
			inter,
			executor.arguments[i],
			argumentType,
			locationRange,
		)
		if err != nil {
			return nil, newError(err, location, codesAndPrograms)
		}
	}

	contractValue, err := inter.GetContractComposite(executor.contractLocation)
	if err != nil {
		return nil, newError(err, location, codesAndPrograms)
	}

	var self interpreter.Value = contractValue

	// prepare invocation
	invocation := interpreter.NewInvocation(
		inter,
		&self,
		nil,
		nil,
		interpreterArguments,
		executor.argumentTypes,
		nil,
		interpreter.LocationRange{
			Location:    context.Location,
			HasPosition: ast.EmptyRange,
		},
	)

	contractMember := contractValue.GetMember(
		inter,
		invocation.LocationRange,
		executor.functionName,
	)

	contractFunction, ok := contractMember.(interpreter.FunctionValue)
	if !ok {
		err := interpreter.NotInvokableError{
			Value: contractFunction,
		}
		return nil, newError(err, location, codesAndPrograms)
	}

	value, err := inter.InvokeFunction(contractFunction, invocation)
	if err != nil {
		return nil, newError(err, location, codesAndPrograms)
	}

	// Write back all stored values, which were actually just cached, back into storage
	err = environment.CommitStorage(inter)
	if err != nil {
		return nil, newError(err, location, codesAndPrograms)
	}

	var exportedValue cadence.Value
	exportedValue, err = ExportValue(value, inter, interpreter.EmptyLocationRange)
	if err != nil {
		return nil, newError(err, location, codesAndPrograms)
	}

	return exportedValue, nil
}

func (executor *interpreterContractFunctionExecutor) convertArgument(
	inter *interpreter.Interpreter,
	argument cadence.Value,
	argumentType sema.Type,
	locationRange interpreter.LocationRange,
) (interpreter.Value, error) {
	environment := executor.environment

	// Convert `Address` arguments to account reference values (`&Account`)
	// if it is the expected argument type,
	// so there is no need for the caller to construct the value

	if addressValue, ok := argument.(cadence.Address); ok {

		if referenceType, ok := argumentType.(*sema.ReferenceType); ok &&
			referenceType.Type == sema.AccountType {

			address := interpreter.NewAddressValue(inter, common.Address(addressValue))

			accountValue := environment.NewAccountValue(inter, address)

			authorization := interpreter.ConvertSemaAccessToStaticAuthorization(
				inter,
				referenceType.Authorization,
			)

			accountReferenceValue := interpreter.NewEphemeralReferenceValue(
				inter,
				authorization,
				accountValue,
				sema.AccountType,
				locationRange,
			)

			return accountReferenceValue, nil
		}
	}

	return ImportValue(
		inter,
		locationRange,
		environment,
		environment.ResolveLocation,
		argument,
		argumentType,
	)
}
