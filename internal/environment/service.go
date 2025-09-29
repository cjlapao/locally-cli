package environment

// import (
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	"sync"

// 	"github.com/cjlapao/locally-cli/internal/appctx"
// 	"github.com/cjlapao/locally-cli/internal/logging"
// 	"github.com/cjlapao/locally-cli/pkg/diagnostics"
// 	"github.com/cjlapao/locally-cli/pkg/interfaces"
// )

// var (
// 	globalEnvironment *Environment
// 	once              sync.Once
// )

// const (
// 	PREFIX string = "${{"
// 	SUFFIX string = "}}"
// )

// // Environment provides a generic environment variable management system
// // that supports vaults and functions for dynamic variable resolution
// type Environment struct {
// 	mu        sync.RWMutex
// 	vaults    map[string]interfaces.EnvironmentVault
// 	functions map[string]interfaces.EnvironmentVariableFunction
// 	variables map[string]map[string]interface{}
// 	isSynced  bool
// }

// // new creates a new Environment instance
// func new() *Environment {
// 	env := &Environment{
// 		vaults:    make(map[string]interfaces.EnvironmentVault),
// 		functions: make(map[string]interfaces.EnvironmentVariableFunction),
// 		variables: make(map[string]map[string]interface{}),
// 		isSynced:  false,
// 	}
// 	return env
// }

// func Initialize() *Environment {
// 	once.Do(func() {
// 		svc := new()

// 		globalEnvironment = svc
// 	})
// 	return globalEnvironment
// }

// // GetInstance returns the global singleton Environment instance
// func GetInstance() *Environment {
// 	if globalEnvironment == nil {
// 		panic("environment not initialized")
// 	}
// 	return globalEnvironment
// }

// // ResetForTesting resets the environment for testing purposes
// // This should only be used in tests
// func ResetForTesting() {
// 	globalEnvironment = new()
// }

// // RegisterVault registers a vault with the environment service
// func (env *Environment) RegisterVault(ctx *appctx.AppContext, vault interfaces.EnvironmentVault) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("register_vault")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name": vault.Name(),
// 	})

// 	env.mu.Lock()
// 	defer env.mu.Unlock()

// 	// Check if vault already exists
// 	if _, exists := env.vaults[vault.Name()]; exists {
// 		diag.AddWarning("VAULT_ALREADY_EXISTS", "Vault already registered", "environment", map[string]interface{}{
// 			"vault_name": vault.Name(),
// 		})
// 		return diag
// 	}

// 	// Try to sync the vault to ensure it's working
// 	_, diag = vault.Sync()
// 	if diag.HasErrors() {
// 		diag.AddError("VAULT_SYNC_FAILED", "Failed to sync vault", "environment", map[string]interface{}{
// 			"vault_name": vault.Name(),
// 		})
// 		return diag
// 	}

// 	// Register the vault
// 	env.vaults[vault.Name()] = vault
// 	env.variables[vault.Name()] = make(map[string]interface{})

// 	diag.AddPathEntry("vault_registered", "environment", map[string]interface{}{
// 		"vault_name":   vault.Name(),
// 		"total_vaults": len(env.vaults),
// 	})

// 	logging.Infof("Registered vault: %s", vault.Name())
// 	return diag
// }

// // RegisterFunction registers a function with the environment service
// func (env *Environment) RegisterFunction(ctx *appctx.AppContext, function interfaces.EnvironmentVariableFunction) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("register_function")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"function_name": function.Name(),
// 	})

// 	env.mu.Lock()
// 	defer env.mu.Unlock()

// 	// Check if function already exists
// 	if _, exists := env.functions[function.Name()]; exists {
// 		diag.AddWarning("FUNCTION_ALREADY_EXISTS", "Function already registered", "environment", map[string]interface{}{
// 			"function_name": function.Name(),
// 		})
// 		return diag
// 	}

// 	// Register the function
// 	env.functions[function.Name()] = function

// 	diag.AddPathEntry("function_registered", "environment", map[string]interface{}{
// 		"function_name":   function.Name(),
// 		"total_functions": len(env.functions),
// 	})

// 	logging.Infof("Registered function: %s", function.Name())
// 	return diag
// }

// // GetVault returns a vault by name
// func (env *Environment) GetVault(ctx *appctx.AppContext, name string) (interfaces.EnvironmentVault, bool) {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	vault, exists := env.vaults[name]
// 	return vault, exists
// }

// // GetFunction returns a function by name
// func (env *Environment) GetFunction(ctx *appctx.AppContext, name string) (interfaces.EnvironmentVariableFunction, bool) {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	function, exists := env.functions[name]
// 	return function, exists
// }

// func (env *Environment) SyncVault(ctx *appctx.AppContext, vaultName string) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("sync_vault")
// 	defer diag.Complete()

// 	vault, exists := env.GetVault(ctx, vaultName)
// 	if !exists {
// 		diag.AddError("VAULT_NOT_FOUND", "Vault not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return diag
// 	}

// 	variables, err := vault.Sync()
// 	if err != nil {
// 		diag.Append(err)
// 	}

// 	env.variables[vaultName] = variables

// 	diag.AddPathEntry("vault_synced", "environment", map[string]interface{}{
// 		"vault_name":      vaultName,
// 		"variables_count": len(variables),
// 	})
// 	return diag
// }

// // ListVaults returns a list of all registered vault names
// func (env *Environment) ListVaults(ctx *appctx.AppContext) []string {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	vaults := make([]string, 0, len(env.vaults))
// 	for name := range env.vaults {
// 		vaults = append(vaults, name)
// 	}
// 	return vaults
// }

// // ListFunctions returns a list of all registered function names
// func (env *Environment) ListFunctions(ctx *appctx.AppContext) []string {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	functions := make([]string, 0, len(env.functions))
// 	for name := range env.functions {
// 		functions = append(functions, name)
// 	}
// 	return functions
// }

// // SetVariable sets a variable in a specific vault
// func (env *Environment) SetVariable(ctx *appctx.AppContext, vaultName, key string, value interface{}) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("set_variable")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 		"value_type": fmt.Sprintf("%T", value),
// 	})

// 	env.mu.Lock()
// 	defer env.mu.Unlock()

// 	// Ensure vault exists
// 	if _, exists := env.vaults[vaultName]; !exists {
// 		diag.AddError("VAULT_NOT_FOUND", "Vault not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return diag
// 	}

// 	// Ensure variables map exists for this vault
// 	if env.variables[vaultName] == nil {
// 		env.variables[vaultName] = make(map[string]interface{})
// 	}

// 	// Set the variable
// 	env.variables[vaultName][strings.ToLower(key)] = value

// 	diag.AddPathEntry("variable_set", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 		"value_type": fmt.Sprintf("%T", value),
// 	})

// 	logging.Infof("Set variable %s.%s: %v", vaultName, key, value)
// 	return diag
// }

// // GetVariable returns a variable from a specific vault
// func (env *Environment) GetVariable(ctx *appctx.AppContext, vaultName, key string) (interface{}, bool) {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	// Check if vault exists
// 	if _, exists := env.vaults[vaultName]; !exists {
// 		return nil, false
// 	}

// 	// Get the variable
// 	value, exists := env.variables[vaultName][strings.ToLower(key)]
// 	return value, exists
// }

// // GetVariableString returns a string variable from a specific vault
// func (env *Environment) GetVariableString(ctx *appctx.AppContext, vaultName, key string) (string, bool) {
// 	value, exists := env.GetVariable(ctx, vaultName, key)
// 	if !exists {
// 		return "", false
// 	}

// 	if str, ok := value.(string); ok {
// 		return str, true
// 	}
// 	return "", false
// }

// // RemoveVariable removes a variable from a specific vault
// func (env *Environment) RemoveVariable(ctx *appctx.AppContext, vaultName, key string) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("remove_variable")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 	})

// 	env.mu.Lock()
// 	defer env.mu.Unlock()

// 	// Check if vault exists
// 	if _, exists := env.vaults[vaultName]; !exists {
// 		diag.AddError(ErrorCodeVaultNotFound, "Vault not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return diag
// 	}

// 	// Check if the variable exists before removing
// 	if _, exists := env.variables[vaultName][strings.ToLower(key)]; !exists {
// 		diag.AddError("VARIABLE_NOT_FOUND", "Variable not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 			"key":        key,
// 		})
// 		return diag
// 	}

// 	// Remove the variable
// 	delete(env.variables[vaultName], strings.ToLower(key))

// 	diag.AddPathEntry("variable_removed", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 	})

// 	logging.Infof("Removed variable %s.%s", vaultName, key)
// 	return diag
// }

// // GetAllVariables returns all variables from a specific vault
// func (env *Environment) GetAllVariables(ctx *appctx.AppContext, vaultName string) (map[string]interface{}, bool) {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	// Check if vault exists
// 	if _, exists := env.vaults[vaultName]; !exists {
// 		return nil, false
// 	}

// 	// Return a copy of the variables
// 	variables := make(map[string]interface{})
// 	for key, value := range env.variables[vaultName] {
// 		variables[key] = value
// 	}

// 	return variables, true
// }

// // ResolveVariable resolves a variable reference like "vault_name.variable_name"
// func (env *Environment) ResolveVariable(ctx *appctx.AppContext, reference string) (interface{}, bool) {
// 	diag := diagnostics.New("resolve_variable")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"reference": reference,
// 	})

// 	// Parse the reference
// 	parts := strings.SplitN(reference, ".", 2)
// 	if len(parts) != 2 {
// 		diag.AddError(ErrorCodeInvalidReference, "Invalid variable reference format", "environment", map[string]interface{}{
// 			"reference":       reference,
// 			"expected_format": "vault_name.variable_name",
// 		})
// 		return nil, false
// 	}

// 	vaultName := parts[0]
// 	variableName := parts[1]

// 	diag.AddPathEntry("parsed_reference", "environment", map[string]interface{}{
// 		"vault_name":    vaultName,
// 		"variable_name": variableName,
// 	})

// 	// Get the variable
// 	value, exists := env.GetVariable(ctx, vaultName, variableName)
// 	if !exists {
// 		diag.AddWarning(ErrorCodeVariableNotFound, "Variable not found", "environment", map[string]interface{}{
// 			"vault_name":    vaultName,
// 			"variable_name": variableName,
// 		})
// 		return nil, false
// 	}

// 	diag.AddPathEntry("variable_resolved", "environment", map[string]interface{}{
// 		"vault_name":    vaultName,
// 		"variable_name": variableName,
// 		"value_type":    fmt.Sprintf("%T", value),
// 	})

// 	return value, true
// }

// // ExecuteFunction executes a registered function with arguments
// func (env *Environment) ExecuteFunction(ctx *appctx.AppContext, functionName string, args ...string) (string, *diagnostics.Diagnostics) {
// 	diag := diagnostics.New("execute_function")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"function_name": functionName,
// 		"args_count":    len(args),
// 	})

// 	env.mu.RLock()
// 	function, exists := env.functions[functionName]
// 	env.mu.RUnlock()

// 	if !exists {
// 		diag.AddError(ErrorCodeFunctionNotFound, "Function not found", "environment", map[string]interface{}{
// 			"function_name": functionName,
// 		})
// 		return "", diag
// 	}

// 	// Execute the function
// 	result, diag := function.Exec("", args...)

// 	diag.AddPathEntry("function_executed", "environment", map[string]interface{}{
// 		"function_name": functionName,
// 		"result_length": len(result),
// 	})

// 	logging.Infof("Executed function %s with args %v: %s", functionName, args, result)
// 	return result, diag
// }

// // Sync synchronizes all vaults or a specific vault
// func (env *Environment) Sync(ctx *appctx.AppContext, vaultNames ...string) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("sync_environment")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_count": len(vaultNames),
// 		"sync_all":    len(vaultNames) == 0,
// 	})

// 	env.mu.Lock()
// 	defer env.mu.Unlock()

// 	// Determine which vaults to sync
// 	vaultsToSync := make([]string, 0)
// 	if len(vaultNames) == 0 {
// 		// Sync all vaults
// 		for name := range env.vaults {
// 			vaultsToSync = append(vaultsToSync, name)
// 		}
// 	} else {
// 		// Sync specific vaults
// 		for _, name := range vaultNames {
// 			if _, exists := env.vaults[name]; exists {
// 				vaultsToSync = append(vaultsToSync, name)
// 			} else {
// 				diag.AddWarning("VAULT_NOT_FOUND", "Vault not found during sync", "environment", map[string]interface{}{
// 					"vault_name": name,
// 				})
// 			}
// 		}
// 	}

// 	diag.AddPathEntry("vaults_identified", "environment", map[string]interface{}{
// 		"vaults_to_sync": vaultsToSync,
// 	})

// 	// Sync each vault
// 	for _, vaultName := range vaultsToSync {
// 		childDiag := diagnostics.New(fmt.Sprintf("sync_vault_%s", vaultName))
// 		defer childDiag.Complete()

// 		vault := env.vaults[vaultName]
// 		variables, vaultDiag := vault.Sync()
// 		if vaultDiag.HasErrors() {
// 			childDiag.Append(vaultDiag)
// 			childDiag.Complete()
// 			// Merge the child diagnostics into the parent
// 			diag.Append(childDiag)
// 			return diag
// 		}

// 		// Update the variables
// 		env.variables[vaultName] = variables

// 		childDiag.AddPathEntry("vault_synced", "environment", map[string]interface{}{
// 			"vault_name":      vaultName,
// 			"variables_count": len(variables),
// 		})

// 		childDiag.Complete()
// 		// Merge the child diagnostics into the parent
// 		diag.Append(childDiag)
// 		logging.Infof("Synced vault %s with %d variables", vaultName, len(variables))
// 	}

// 	env.isSynced = true

// 	diag.AddPathEntry("sync_completed", "environment", map[string]interface{}{
// 		"total_vaults_synced": len(vaultsToSync),
// 	})

// 	logging.Infof("Environment sync completed for %d vaults", len(vaultsToSync))
// 	return diag
// }

// // IsSynced returns whether the environment has been synced
// func (env *Environment) IsSynced() bool {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()
// 	return env.isSynced
// }

// // Replace replaces variable references in a string
// func (env *Environment) Replace(ctx *appctx.AppContext, source string) string {
// 	diag := diagnostics.New("replace_variables")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"source_length": len(source),
// 	})

// 	result := source
// 	start := 0

// 	for {
// 		// Find the next variable reference
// 		prefixIndex := strings.Index(result[start:], PREFIX)
// 		if prefixIndex == -1 {
// 			break
// 		}
// 		prefixIndex += start

// 		suffixIndex := strings.Index(result[prefixIndex:], SUFFIX)
// 		if suffixIndex == -1 {
// 			break
// 		}
// 		suffixIndex += prefixIndex

// 		// Extract the variable reference
// 		variableRef := result[prefixIndex+len(PREFIX) : suffixIndex]
// 		variableRef = strings.TrimSpace(variableRef)

// 		diag.AddPathEntry("variable_found", "environment", map[string]interface{}{
// 			"variable_ref": variableRef,
// 		})

// 		// Check if it's a function call
// 		if strings.Contains(variableRef, "(") && strings.Contains(variableRef, ")") {
// 			replacement, diag := env.processFunctionCall(ctx, variableRef, diag)
// 			if diag.HasErrors() {
// 				diag.AddWarning("FUNCTION_EXECUTION_FAILED", "Function execution failed", "environment", map[string]interface{}{
// 					"variable_ref": variableRef,
// 					"error":        diag.GetErrors()[0].Message,
// 				})
// 				// Keep the original reference if function execution fails
// 				start = suffixIndex + len(SUFFIX)
// 				continue
// 			}
// 			result = result[:prefixIndex] + replacement + result[suffixIndex+len(SUFFIX):]
// 		} else if strings.Contains(variableRef, ":") {
// 			// Handle colon-separated function calls (e.g., "test-function:test-value:arg1:arg2")
// 			parts := strings.Split(variableRef, ":")
// 			if len(parts) >= 2 {
// 				functionName := parts[0]
// 				args := parts[1:]

// 				// Check if the function exists
// 				env.mu.RLock()
// 				_, exists := env.functions[functionName]
// 				env.mu.RUnlock()

// 				if exists {
// 					// Execute the function
// 					replacement, diag := env.ExecuteFunction(ctx, functionName, args...)
// 					if diag.HasErrors() {
// 						diag.AddWarning("FUNCTION_EXECUTION_FAILED", "Function execution failed", "environment", map[string]interface{}{
// 							"variable_ref": variableRef,
// 							"error":        diag.GetErrors()[0].Message,
// 						})
// 						// Keep the original reference if function execution fails
// 						start = suffixIndex + len(SUFFIX)
// 						continue
// 					}
// 					result = result[:prefixIndex] + replacement + result[suffixIndex+len(SUFFIX):]
// 				} else {
// 					// Function not found, keep original reference
// 					start = suffixIndex + len(SUFFIX)
// 					continue
// 				}
// 			} else {
// 				// Invalid format, keep original reference
// 				start = suffixIndex + len(SUFFIX)
// 				continue
// 			}
// 		} else {
// 			// It's a regular variable reference
// 			value, found := env.ResolveVariable(ctx, variableRef)
// 			if found {
// 				replacement := fmt.Sprintf("%v", value)
// 				result = result[:prefixIndex] + replacement + result[suffixIndex+len(SUFFIX):]
// 				diag.AddPathEntry("variable_replaced", "environment", map[string]interface{}{
// 					"variable_ref": variableRef,
// 					"replacement":  replacement,
// 				})
// 			} else {
// 				diag.AddWarning("VARIABLE_NOT_FOUND", "Variable not found during replacement", "environment", map[string]interface{}{
// 					"variable_ref": variableRef,
// 				})
// 				// Keep the original reference if variable not found
// 				start = suffixIndex + len(SUFFIX)
// 				continue
// 			}
// 		}

// 		// Reset start position for next iteration
// 		start = 0
// 	}

// 	diag.AddPathEntry("replace_completed", "environment", map[string]interface{}{
// 		"result_length": len(result),
// 	})

// 	return result
// }

// // processFunctionCall processes a function call like "function_name(arg1, arg2)"
// func (env *Environment) processFunctionCall(ctx *appctx.AppContext, functionCall string, diag *diagnostics.Diagnostics) (string, *diagnostics.Diagnostics) {
// 	diag.AddPathEntry("process_function_call", "environment", map[string]interface{}{
// 		"function_call": functionCall,
// 	})

// 	// Parse function name and arguments
// 	openParen := strings.Index(functionCall, "(")
// 	closeParen := strings.LastIndex(functionCall, ")")

// 	if openParen == -1 || closeParen == -1 || closeParen <= openParen {
// 		diag.AddError("INVALID_FUNCTION_CALL", "Invalid function call format", "environment", map[string]interface{}{
// 			"function_call": functionCall,
// 		})
// 		return "", diag
// 	}

// 	functionName := strings.TrimSpace(functionCall[:openParen])
// 	argsString := strings.TrimSpace(functionCall[openParen+1 : closeParen])

// 	// Parse arguments
// 	var args []string
// 	if argsString != "" {
// 		args = strings.Split(argsString, ",")
// 		for i, arg := range args {
// 			args[i] = strings.TrimSpace(arg)
// 		}
// 	}

// 	// Execute the function
// 	result, diag := env.ExecuteFunction(ctx, functionName, args...)
// 	return result, diag
// }

// // GetStatus returns the current status of the environment
// func (env *Environment) GetStatus(ctx *appctx.AppContext) map[string]interface{} {
// 	env.mu.RLock()
// 	defer env.mu.RUnlock()

// 	status := map[string]interface{}{
// 		"synced":    env.isSynced,
// 		"vaults":    make([]string, 0),
// 		"functions": make([]string, 0),
// 		"variables": make(map[string]int),
// 	}

// 	// Add vault names
// 	for name := range env.vaults {
// 		status["vaults"] = append(status["vaults"].([]string), name)
// 	}

// 	// Add function names
// 	for name := range env.functions {
// 		status["functions"] = append(status["functions"].([]string), name)
// 	}

// 	// Add variable counts per vault
// 	for vaultName, variables := range env.variables {
// 		status["variables"].(map[string]int)[vaultName] = len(variables)
// 	}

// 	return status
// }

// // Add adds a key-value pair to a specific vault
// func (env *Environment) Add(ctx *appctx.AppContext, vaultName, key string, value interface{}) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("add_to_vault")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 		"value_type": fmt.Sprintf("%T", value),
// 	})

// 	env.mu.RLock()
// 	vault, exists := env.vaults[vaultName]
// 	env.mu.RUnlock()

// 	if !exists {
// 		diag.AddError(ErrorCodeVaultNotFound, "Vault not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return diag
// 	}

// 	// Delegate to the vault
// 	err := vault.Set(key, value)
// 	if err != nil {
// 		diag.Append(err)
// 		return diag
// 	}

// 	diag.AddPathEntry("value_added", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 	})

// 	logging.Infof("Added %s.%s: %v", vaultName, key, value)
// 	return diag
// }

// // Remove removes a key from a specific vault
// func (env *Environment) Remove(ctx *appctx.AppContext, vaultName, key string) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("remove_from_vault")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 	})

// 	env.mu.RLock()
// 	vault, exists := env.vaults[vaultName]
// 	env.mu.RUnlock()

// 	if !exists {
// 		diag.AddError(ErrorCodeVaultNotFound, "Vault not found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return diag
// 	}

// 	// Delegate to the vault
// 	err := vault.Remove(key)
// 	if err != nil {
// 		diag.Append(err)
// 		return diag
// 	}

// 	diag.AddPathEntry("value_removed", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 	})

// 	logging.Infof("Removed %s.%s", vaultName, key)
// 	return diag
// }

// // Get gets a value from a specific vault with a default value
// func (env *Environment) Get(ctx *appctx.AppContext, vaultName, key string, defaultValue interface{}) (interface{}, *diagnostics.Diagnostics) {
// 	diag := diagnostics.New("get_from_vault")
// 	defer diag.Complete()

// 	diag.AddPathEntry("start", "environment", map[string]interface{}{
// 		"vault_name":    vaultName,
// 		"key":           key,
// 		"default_value": defaultValue,
// 	})

// 	env.mu.RLock()
// 	vault, exists := env.vaults[vaultName]
// 	env.mu.RUnlock()

// 	if !exists {
// 		diag.AddWarning(ErrorCodeVaultNotFound, "Vault not found, returning default", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 		})
// 		return defaultValue, diag
// 	}

// 	// First check the environment's internal variables map
// 	env.mu.RLock()
// 	if env.variables[vaultName] != nil {
// 		if value, found := env.variables[vaultName][strings.ToLower(key)]; found {
// 			env.mu.RUnlock()
// 			diag.AddPathEntry("value_found_in_env", "environment", map[string]interface{}{
// 				"vault_name": vaultName,
// 				"key":        key,
// 				"value_type": fmt.Sprintf("%T", value),
// 			})
// 			return value, diag
// 		}
// 	}
// 	env.mu.RUnlock()

// 	// If not found in environment, delegate to the vault
// 	value, found := vault.Get(key)
// 	if !found {
// 		diag.AddPathEntry("key_not_found", "environment", map[string]interface{}{
// 			"vault_name": vaultName,
// 			"key":        key,
// 		})
// 		return defaultValue, diag
// 	}

// 	diag.AddPathEntry("value_found_in_vault", "environment", map[string]interface{}{
// 		"vault_name": vaultName,
// 		"key":        key,
// 		"value_type": fmt.Sprintf("%T", value),
// 	})

// 	return value, diag
// }

// // GetString gets a string value from a specific vault with a default value
// func (env *Environment) GetString(ctx *appctx.AppContext, vaultName, key string, defaultValue string) (string, *diagnostics.Diagnostics) {
// 	value, diag := env.Get(ctx, vaultName, key, defaultValue)
// 	if value == nil {
// 		return defaultValue, diag
// 	}

// 	switch v := value.(type) {
// 	case string:
// 		return v, diag
// 	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
// 		return fmt.Sprintf("%v", v), diag
// 	default:
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert value to string", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       value,
// 			"value_type":  fmt.Sprintf("%T", value),
// 			"target_type": "string",
// 		})
// 		return defaultValue, diag
// 	}
// }

// // GetBool gets a boolean value from a specific vault with a default value
// func (env *Environment) GetBool(ctx *appctx.AppContext, vaultName, key string, defaultValue bool) (bool, *diagnostics.Diagnostics) {
// 	value, diag := env.Get(ctx, vaultName, key, defaultValue)
// 	if value == nil {
// 		return defaultValue, diag
// 	}

// 	switch v := value.(type) {
// 	case bool:
// 		return v, diag
// 	case string:
// 		switch strings.ToLower(v) {
// 		case "true", "1", "yes", "on":
// 			return true, diag
// 		case "false", "0", "no", "off":
// 			return false, diag
// 		default:
// 			diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert string to bool", "environment", map[string]interface{}{
// 				"vault_name":  vaultName,
// 				"key":         key,
// 				"value":       v,
// 				"target_type": "bool",
// 			})
// 			return defaultValue, diag
// 		}
// 	case int, int8, int16, int32, int64:
// 		return v != 0, diag
// 	case uint, uint8, uint16, uint32, uint64:
// 		return v != 0, diag
// 	case float32, float64:
// 		return v != 0, diag
// 	default:
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert value to bool", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       value,
// 			"value_type":  fmt.Sprintf("%T", value),
// 			"target_type": "bool",
// 		})
// 		return defaultValue, diag
// 	}
// }

// // GetInt gets an integer value from a specific vault with a default value
// func (env *Environment) GetInt(ctx *appctx.AppContext, vaultName, key string, defaultValue int) (int, *diagnostics.Diagnostics) {
// 	value, diag := env.Get(ctx, vaultName, key, defaultValue)
// 	if value == nil {
// 		return defaultValue, diag
// 	}

// 	switch v := value.(type) {
// 	case int:
// 		return v, diag
// 	case int8:
// 		return int(v), diag
// 	case int16:
// 		return int(v), diag
// 	case int32:
// 		return int(v), diag
// 	case int64:
// 		return int(v), diag
// 	case uint:
// 		return int(v), diag
// 	case uint8:
// 		return int(v), diag
// 	case uint16:
// 		return int(v), diag
// 	case uint32:
// 		return int(v), diag
// 	case uint64:
// 		return int(v), diag
// 	case float32:
// 		return int(v), diag
// 	case float64:
// 		return int(v), diag
// 	case string:
// 		if i, err := strconv.Atoi(v); err == nil {
// 			return i, diag
// 		}
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert string to int", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       v,
// 			"target_type": "int",
// 		})
// 		return defaultValue, diag
// 	case bool:
// 		if v {
// 			return 1, diag
// 		}
// 		return 0, diag
// 	default:
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert value to int", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       value,
// 			"value_type":  fmt.Sprintf("%T", value),
// 			"target_type": "int",
// 		})
// 		return defaultValue, diag
// 	}
// }

// // GetFloat gets a float64 value from a specific vault with a default value
// func (env *Environment) GetFloat(ctx *appctx.AppContext, vaultName, key string, defaultValue float64) (float64, *diagnostics.Diagnostics) {
// 	value, diag := env.Get(ctx, vaultName, key, defaultValue)
// 	if value == nil {
// 		return defaultValue, diag
// 	}

// 	switch v := value.(type) {
// 	case float64:
// 		return v, diag
// 	case float32:
// 		return float64(v), diag
// 	case int:
// 		return float64(v), diag
// 	case int8:
// 		return float64(v), diag
// 	case int16:
// 		return float64(v), diag
// 	case int32:
// 		return float64(v), diag
// 	case int64:
// 		return float64(v), diag
// 	case uint:
// 		return float64(v), diag
// 	case uint8:
// 		return float64(v), diag
// 	case uint16:
// 		return float64(v), diag
// 	case uint32:
// 		return float64(v), diag
// 	case uint64:
// 		return float64(v), diag
// 	case string:
// 		if f, err := strconv.ParseFloat(v, 64); err == nil {
// 			return f, diag
// 		}
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert string to float64", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       v,
// 			"target_type": "float64",
// 		})
// 		return defaultValue, diag
// 	case bool:
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Cannot convert bool to float64", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       v,
// 			"target_type": "float64",
// 		})
// 		return defaultValue, diag
// 	default:
// 		diag.AddWarning(ErrorCodeTypeConversionError, "Failed to convert value to float64", "environment", map[string]interface{}{
// 			"vault_name":  vaultName,
// 			"key":         key,
// 			"value":       value,
// 			"value_type":  fmt.Sprintf("%T", value),
// 			"target_type": "float64",
// 		})
// 		return defaultValue, diag
// 	}
// }
