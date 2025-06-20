package config

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
)

var errNotStructPointer = errors.New("parsing target must be a pointer to struct")

// LoadEnv loads configuration data from environment variables, whether it's from files or from OS.
// If no files is provided, by default this function will load from
// file .env or from OS environment variables.
// If the same key in a file also exists in OS env, the key value
// will be overriden by the OS env value
func LoadEnv(v any, files ...string) error {
	m, err := godotenv.Read(files...)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return errNotStructPointer
	}

	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errNotStructPointer
	}

	if !rv.CanSet() {
		return errNotStructPointer
	}

	err = parseStruct(environmentVariables(m), rv, "", "")

	return err
}

func createParsingError(envk string, err error) error {
	return fmt.Errorf("error on parsing env %s: %w", envk, err)
}

type environmentVariables map[string]string

func (envVars environmentVariables) get(envk, defaultVal string) (string, bool) {
	val, exists := os.LookupEnv(envk)
	if !exists {
		val, exists = envVars[envk]
		if !exists {
			val = defaultVal
			exists = defaultVal != ""
		}
	}

	return val, exists
}

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func parseTextUnmarshaler(envVars environmentVariables, rv reflect.Value, envk, defaultVal string) (bool, error) {
	rt := rv.Type()
	rvCopy := rv
	mutateToPointer := false
	if rt.Kind() != reflect.Ptr {
		newRt := reflect.PointerTo(rt)
		rt = newRt

		rvCopy = reflect.New(rt.Elem())
		mutateToPointer = true
	}

	if !rt.Implements(textUnmarshalerType) {
		return false, nil
	}

	unmarshalTextMethod := rvCopy.MethodByName("UnmarshalText")
	if !unmarshalTextMethod.IsValid() {
		return false, errors.New("invalid data type")
	}

	enval, ok := envVars.get(envk, defaultVal)
	if !ok && defaultVal == "" {
		return true, nil
	}

	// Call UnmarshalText([]byte(enval))
	results := unmarshalTextMethod.Call([]reflect.Value{
		reflect.ValueOf([]byte(enval)),
	})

	// Check the returned error from calling UnmarshalText([]byte(enval))
	errRv := results[0]
	if errRv.IsValid() {
		if !errRv.IsZero() {
			return false, errRv.Elem().Interface().(error)
		}
	}

	if mutateToPointer {
		rv.Set(rvCopy.Elem())
	}

	return true, nil
}

func parseStruct(envVars environmentVariables, rv reflect.Value, envk, defaultVal string) error {
	// A struct that has environment variable attributed to it
	// must implements encoding.TextUnmarshaler so that the
	// text representation of this struct can be decoded
	if envk != "" {
		ok, err := parseTextUnmarshaler(envVars, rv, envk, defaultVal)
		if err != nil {
			return err
		}

		if !ok {
			return err
		}

		return nil
	}

	for i := range rv.NumField() {
		sf := rv.Type().Field(i)
		sfv := rv.Field(i)

		// Get env key
		envk := sf.Tag.Get("env")

		// Get env default value
		defaultVal := sf.Tag.Get("default")
		// Try to parse with parseTextUnmarshaler first
		ok, err := parseTextUnmarshaler(envVars, sfv, envk, defaultVal)
		if err != nil {
			return createParsingError(envk, err)
		}

		if ok {
			continue
		}

		if _, err := parseAny(envVars, sfv, envk, defaultVal); err != nil {
			return createParsingError(envk, err)
		}
	}

	return nil
}

func parsePointer(envVars environmentVariables, rv reflect.Value, envk, defaultVal string) error {
	ok, err := parseTextUnmarshaler(envVars, rv, envk, defaultVal)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	rvet := rv.Type().Elem()
	newRv := reflect.New(rvet)

	exists, err := parseAny(envVars, newRv.Elem(), envk, defaultVal)
	if err != nil {
		return err
	}

	if exists {
		rv.Set(newRv)
	}

	return nil
}

func parseAny(envVars environmentVariables, rv reflect.Value, envk, defaultVal string) (bool, error) {
	enval, envExists := envVars.get(envk, defaultVal)

	switch rv.Kind() {
	case reflect.Struct:
		if err := parseStruct(envVars, rv, envk, defaultVal); err != nil {
			return envExists, err
		}

	case reflect.Pointer:
		if err := parsePointer(envVars, rv, envk, defaultVal); err != nil {
			return envExists, err
		}

	case reflect.Interface:
		if envExists {
			rv.Set(reflect.ValueOf(enval))
		}

	case reflect.String:
		enval, _ := envVars.get(envk, defaultVal)
		rv.SetString(enval)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if enval != "" {
			i64, err := strconv.ParseInt(enval, 10, 64)
			if err != nil {
				return envExists, err
			}

			rv.SetInt(i64)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if enval != "" {
			ui64, err := strconv.ParseUint(enval, 10, 64)
			if err != nil {
				return envExists, err
			}

			rv.SetUint(ui64)
		}

	case reflect.Float32, reflect.Float64:
		if enval != "" {
			f64, err := strconv.ParseFloat(enval, 64)
			if err != nil {
				return envExists, err
			}

			rv.SetFloat(f64)
		}

	case reflect.Bool:
		if enval != "" {
			switch enval {
			case "1", "true":
				rv.SetBool(true)

			case "0", "false":
				rv.SetBool(false)

			default:
				return envExists, errors.New("string is not a valid boolean representation")
			}
		}

	default:
		return envExists, fmt.Errorf("data type %s is not supported", rv.Kind().String())
	}

	return envExists, nil
}
