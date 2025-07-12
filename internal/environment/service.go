package environment

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cjlapao/locally-cli/internal/environment/functions/random"
	env_interfaces "github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/interfaces"
	"github.com/cjlapao/locally-cli/internal/tools"
	"github.com/cjlapao/locally-cli/internal/vaults/backend_vault"
	"github.com/cjlapao/locally-cli/internal/vaults/config_vault"
	"github.com/cjlapao/locally-cli/internal/vaults/credentials_vault"

	"github.com/cjlapao/locally-cli/internal/vaults/global_vault"
	"github.com/cjlapao/locally-cli/internal/vaults/keyvault_vault"
	"github.com/cjlapao/locally-cli/internal/vaults/terraform_vault"

	"github.com/cjlapao/common-go/guard"
)

var globalEnvironment *Environment

const (
	PREFIX string = "${{"
	SUFFIX string = "}}"
)

type Environment struct {
	isInitialized bool
	variables     map[string]map[string]interface{}
	vaults        []interfaces.EnvironmentVault
	isSync        bool
	functions     []env_interfaces.VariableFunction
}

func New() *Environment {
	svc := Environment{
		variables:     make(map[string]map[string]interface{}),
		vaults:        make([]interfaces.EnvironmentVault, 0),
		functions:     make([]env_interfaces.VariableFunction, 0),
		isSync:        false,
		isInitialized: false,
	}

	// Adding system environment variables
	svc.addEnvironment("API_PREFIX", "api")

	globalEnvironment = &svc
	return globalEnvironment
}

func (env *Environment) IsInitialized() bool {
	return env.isInitialized
}

func (env *Environment) Initialize() {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("unable to sync environment: %v", r)
			notify.Error(err.Error())
			env.isInitialized = false
		}
	}()

	if len(env.vaults) > 0 {
		env.vaults = make([]interfaces.EnvironmentVault, 0)
	}
	if len(env.functions) > 0 {
		env.functions = make([]env_interfaces.VariableFunction, 0)
	}

	// Adding environment vaults
	env.vaults = append(env.vaults, config_vault.New())
	env.vaults = append(env.vaults, credentials_vault.New())
	env.vaults = append(env.vaults, backend_vault.New())
	env.vaults = append(env.vaults, global_vault.New())
	env.vaults = append(env.vaults, terraform_vault.New())
	env.vaults = append(env.vaults, keyvault_vault.New())

	// Adding environment functions
	env.functions = append(env.functions, random.RandomValueFunction{})

	// Reading the values at the beginning
	if err := env.Sync(); err != nil {
		env.isInitialized = false
	} else {
		env.isInitialized = true
	}
}

func Get() *Environment {
	if globalEnvironment != nil {
		if !globalEnvironment.isInitialized {
			globalEnvironment.Initialize()
		}
		return globalEnvironment
	}

	return New()
}

func (env *Environment) Register(vault interfaces.EnvironmentVault) error {
	exists := false
	for _, envVaults := range env.vaults {
		if strings.EqualFold(envVaults.Name(), vault.Name()) {
			exists = true
			break
		}
	}

	if !exists {
		env.vaults = append(env.vaults, vault)
		_, err := vault.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Environment) Add(vault, key string, value interface{}) error {
	key = strings.ToLower(key)

	if err := guard.EmptyOrNil(vault); err != nil {
		notify.Error(err.Error())
		return err
	}

	if err := guard.EmptyOrNil(key); err != nil {
		notify.Error(err.Error())
		return err
	}

	if err := guard.EmptyOrNil(value); err != nil {
		notify.Error(err.Error())
		return err
	}

	if _, ok := env.variables[vault]; !ok {
		env.variables[vault] = make(map[string]interface{})
	}

	env.variables[vault][key] = value

	notify.Debug("%s.%s: %v", vault, key, fmt.Sprintf("%v", value))
	return nil
}

func (env *Environment) Remove(vault, key string) error {
	key = strings.ToLower(key)
	if err := guard.EmptyOrNil(env.variables[vault]); err != nil {
		notify.Error(err.Error())
		return err
	}

	if err := guard.EmptyOrNil(env.variables[vault][key]); err != nil {
		notify.Error(err.Error())
		return err
	}

	delete(env.variables[vault], key)

	if len(env.variables[vault]) == 0 {
		delete(env.variables, vault)
	}

	return nil
}

func (env *Environment) Get(vault, key string) interface{} {
	key = strings.ToLower(key)
	if err := guard.EmptyOrNil(env.variables[vault]); err != nil {
		notify.Error(err.Error())
		return err
	}

	if _, ok := env.variables[vault]; ok {
		if _, ok := env.variables[vault][key]; ok {
			return env.variables[vault][key]
		}
	}

	return nil
}

func (env *Environment) GetAll(vault string) ([]string, error) {
	result := make([]string, 0)
	if err := guard.EmptyOrNil(env.variables[vault]); err != nil {
		notify.Error(err.Error())
		return result, err
	}

	if _, ok := env.variables[vault]; ok {
		for key, val := range env.variables[vault] {
			result = append(result, fmt.Sprintf("%s: %v", key, val))
		}
	}

	return result, nil
}

func (env *Environment) GetString(vault, key string) string {
	value := env.Get(vault, key)
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

func (env *Environment) GetBool(vault, key string) bool {
	value := env.Get(vault, key)
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return strings.EqualFold("true", v)
	case int:
		return v == 1
	case bool:
		return v
	default:
		return false
	}
}

func (env *Environment) extract(source string) []string {
	result := make([]string, 0)
	for {
		source = strings.TrimSpace(source)
		if source == "" {
			break
		}

		initialPos := strings.Index(source, PREFIX)
		finalPos := strings.Index(source, SUFFIX)
		if initialPos == -1 || finalPos == -1 {
			result = append(result, source)
			break
		}

		if finalPos <= initialPos {
			result = append(result, source)
			break
		}

		if initialPos > 0 {
			fragment := source[:initialPos]
			result = append(result, fragment)
		}

		fragment := source[initialPos : finalPos+len(SUFFIX)]
		result = append(result, fragment)
		source = source[finalPos+len(SUFFIX):]
	}

	return result
}

func (env *Environment) Replace(source string) string {
	fragments := env.extract(source)
	replacedFragments := env.replaceFragments(fragments)

	return strings.Join(replacedFragments, "")
}

func (env *Environment) replaceFragments(fragments []string) []string {
	result := make([]string, len(fragments))
	for _, fragment := range fragments {
		if err := guard.EmptyOrNil(fragment); err != nil {
			continue
		}

		if !strings.HasPrefix(fragment, PREFIX) {
			result = append(result, fragment)
			continue
		}

		if !strings.HasSuffix(fragment, SUFFIX) {
			result = append(result, fragment)
			continue
		}

		vault := ""
		key := ""

		cleaned := strings.TrimSpace(strings.TrimPrefix(fragment, PREFIX))
		cleaned = strings.TrimSpace(strings.TrimSuffix(cleaned, SUFFIX))

		parts := strings.Split(cleaned, ".")
		if len(parts) < 2 {
			key = cleaned
		} else {
			vault = parts[0]
			key = ""
			if len(parts) == 2 {
				key = parts[1]
			} else {
				key = strings.Join(parts[1:], ".")
			}
		}

		functions := make([]string, 0)
		functionsParts := strings.Split(key, "|")
		if len(functionsParts) > 0 {
			notify.Debug("found function")
			key = strings.TrimSpace(functionsParts[0])
			if len(functionsParts) > 1 {
				functions = functionsParts[1:]
			}
		}

		for _, function := range functions {
			notify.Debug(function)
		}

		if vault == "" && len(functions) > 0 {
			for _, function := range functions {
				funcArgs := env.extractFunctionArgs(function)
				notify.Debug("Trying to execute with args %s", strings.Join(funcArgs, "."))
				key = env.execute(fmt.Sprintf("%v", key), funcArgs...)
			}

			result = append(result, key)
			continue
		}

		if strings.EqualFold(vault, "ems") && strings.EqualFold(key, "api.key") {
			emsTool := tools.EmsServiceTool{}
			key, _ := emsTool.GenerateEmsApiKeyHeader(env.GetString("keyvault", "global--environment--apikey"))
			if key != "" {
				result = append(result, fmt.Sprintf("%v", key))
			}
		} else {
			value := env.Get(vault, key)
			if value == nil {
				notify.Debug("Key %s was not found in vault %s", key, vault)

				for _, function := range functions {
					funcArgs := env.extractFunctionArgs(function)
					notify.Debug("Trying to execute with args %s", strings.Join(funcArgs, "."))
					value = env.execute(fmt.Sprintf("%v", value), funcArgs...)
				}

				result = append(result, fragment)
				continue
			}

			if strings.ContainsAny(fmt.Sprintf("%v", value), PREFIX) && strings.ContainsAny(fmt.Sprintf("%v", value), SUFFIX) {
				notify.Debug("found nested variable %s", fmt.Sprintf("%v", value))
				// resetting functions that needs applying as it is a nested value
				functions = make([]string, 0)
				value = env.Replace(fmt.Sprintf("%v", value))
			}

			for _, function := range functions {
				funcArgs := env.extractFunctionArgs(function)
				notify.Debug("Trying to execute with args %s", strings.Join(funcArgs, "."))
				value = env.execute(fmt.Sprintf("%v", value), funcArgs...)
			}

			notify.Debug("Key %s was found in vault %s and replaced with %v", key, vault, value)
			result = append(result, fmt.Sprintf("%v", value))
		}
	}

	return result
}

func (env *Environment) Sync() error {
	return env.vaultSync("", false)
}

func (env *Environment) SyncVault(vaultName string) error {
	return env.vaultSync(vaultName, false)
}

func (env *Environment) Refresh() error {
	return env.vaultSync("", true)
}

func (env *Environment) RefreshVault(vaultName string) error {
	return env.vaultSync(vaultName, true)
}

func (env *Environment) vaultSync(vaultName string, force bool) error {
	if env.isSync && !force {
		notify.Debug("Environment already synced, ignoring")
		return nil
	}

	for _, vaultInterface := range env.vaults {
		shouldSync := false
		if vaultName == "" {
			shouldSync = true
		} else {
			if strings.EqualFold(vaultInterface.Name(), vaultName) {
				shouldSync = true
			}
		}

		if shouldSync {
			notify.Debug("Starting to sync vault %s", vaultInterface.Name())
			kv, err := vaultInterface.Sync()
			if err != nil {
				return err
			}
			env.variables[vaultInterface.Name()] = kv
		} else {
			notify.Debug("Ignoring the sync of vault %s, not the requested one", vaultInterface.Name())
		}
	}

	env.isSync = true
	return nil
}

func (env *Environment) extractFunctionArgs(function string) []string {
	function = strings.TrimSpace(function)
	return strings.Split(function, " ")
}

func (env *Environment) execute(value string, args ...string) string {
	for _, function := range env.functions {
		executer := function.New()
		value = executer.Exec(value, args...)
	}
	return value
}

func (env *Environment) addEnvironment(key, value string) error {
	if err := guard.EmptyOrNil(key); err != nil {
		return err
	}
	if err := env.Add("env", key, value); err != nil {
		return err
	}

	return nil
}

// func (env *Environment) Decode(source interface{}) error {
// 	var elem reflect.Value
// 	var sourceType = reflect.TypeOf(source)
// 	if sourceType.Kind() != reflect.Ptr {
// 		fmt.Printf("this is not a pointer %v\n", source)
// 		elem = reflect.ValueOf(source)
// 	} else {
// 		elem = reflect.ValueOf(source).Elem()
// 	}

// 	env.crawler(elem)

// 	return nil
// }

// func (env *Environment) crawler(value reflect.Value) {

// 	fmt.Printf("field is %v: %v\n", value, value.Kind())

// 	switch value.Kind() {
// 	case reflect.String:
// 		if value.CanSet() {
// 			fmt.Printf("%v is a string and changing it to %s\n", value, env.Replace(value.String()))
// 		} else {
// 			// result := env.Replace(value.String())
// 			scatch := reflect.New(value.Type().Elem())

// 			value.Set(scatch)
// 			fmt.Printf("%v cannot be set\n", value)
// 		}
// 	case reflect.Interface:
// 		env.crawler(value.Elem())
// 	case reflect.Array, reflect.Slice:
// 		for i := 0; i < value.Len(); i++ {
// 			env.crawler(value.Index(i))
// 		}
// 	case reflect.Map:
// 		for _, k := range value.MapKeys() {
// 			env.crawler(value.MapIndex(k))
// 		}
// 	case reflect.Pointer:
// 		env.crawler(value.Elem())
// 	case reflect.Struct:
// 		for i := 0; i < value.NumField(); i++ {
// 			f := value.Field(i)
// 			t := f.Kind()
// 			fmt.Printf("Found field %v: %v\n", f, t)
// 			// switch t {
// 			// case reflect.String:
// 			// 	fmt.Printf("is a string and changing it to %s\n", f)
// 			// case reflect.Array, reflect.Slice:
// 			// 	for i := 0; i < f.Len(); i++ {
// 			// 		env.mapd(f.Index(i))
// 			// 	}
// 			// case reflect.Map:
// 			// 	for _, k := range f.MapKeys() {
// 			// 		env.mapd(f.MapIndex(k))
// 			// 	}
// 			// case reflect.Struct, reflect.Pointer:
// 			// 	fmt.Printf("is not a string %s\n", f)
// 			// 	env.mapd(f)

// 			// }
// 			env.crawler(f)
// 		}
// 	}
// }

// func (env *Environment) marshalValue(value interface{}, level int, indent string) string {
// 	switch t := value.(type) {
// 	case string:
// 		t = env.Replace(t)
// 		notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", t))

// 	case map[string]interface{}:
// 		for _, val := range value.(map[string]interface{}) {
// 			val = env.marshalValue(val, level+1, indent)
// 		}
// 	case []interface{}:
// 		notify.Debug("Found variable of type %s with value %s", fmt.Sprintf("%T", t), fmt.Sprintf("%v", value))
// 		for _, val := range value.([]interface{}) {
// 			val = env.marshalValue(val, level+1, indent)
// 		}
// 	default:
// 		return "not supported"
// 	}

// 	return ""
// }
