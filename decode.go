package envparser

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var log = slog.Default()

func SetLogger(n *slog.Logger) {
	if n != nil {
		log = n
	}
}

func SetUpFlags(splitter string, p any) error {
	return setUpFlags("", "", splitter, p)
}

func setUpFlags(prefix, env, splitter string, p any) error {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr {
		log.Error("expected a pointer to a struct", "type", p)
		return fmt.Errorf("expected a pointer to a struct, got %T", p)
	}
	// Dereference the pointer to get the struct value
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		log.Error("expected a pointer to a struct", "type", v.Kind())
		return fmt.Errorf("expected a pointer to a struct, got pointer to %v", v.Kind())
	}

	t := v.Type()

	// loop over the fields of the struct
	fields := reflect.VisibleFields(t)
	for _, field := range fields {
		fieldInfo := v.FieldByIndex(field.Index)
		value := fieldInfo.Addr().Interface()

		fieldsEnv := field.Tag.Get("env")
		// if env is not set then use the parent env + the field env
		if env != "" && fieldsEnv != "" {
			fieldsEnv = env + splitter + fieldsEnv
		}

		rawFlagTag, hasField := field.Tag.Lookup("flag")
		if !hasField || rawFlagTag == "-" {
			continue
		}

		flagDefaultValue := ""
		flagUsage := ""

		tagData := strings.Split(strings.TrimSpace(rawFlagTag), ",")
		flagKey := tagData[0]
		if prefix != "" {
			flagKey = prefix + splitter + flagKey
		}
		log := log.With(
			"name", field.Name,
			"type", field.Type,
			"flag-key", flagKey,
			"value", fmt.Sprintf("%+v", fieldInfo.Interface()),
			"env-key", fieldsEnv,
		)

		exampleTag := `tag should look like flag:"flag-name, default:some value, usage:help info for this flag"`
		for _, s := range tagData[1:] {
			k := strings.SplitN(s, ":", 2)
			if len(k) != 2 {
				log.Error("error: unsupported tag", "tag", s, "field", field.Name, "fix", exampleTag)
				return fmt.Errorf("error: unsupported tag %q for field %q\n tag should look `"+exampleTag+"`", s, field.Name)
			}

			switch strings.TrimSpace(k[0]) {
			case "default":
				flagDefaultValue = k[1]
			case "usage":
				flagUsage = k[1]
			default:
				log.Error("error: unsupported tag", "tag", s, "field", field.Name, "fix", exampleTag)
				return fmt.Errorf("error: unsupported tag %q for field %q\n", s, field.Name)
			}
		}

		supportedTypes := []string{"string", "int", "int64", "uint", "uint64", "bool", "struct"}
		log = log.With("flag-usage", flagUsage, "flag-default", flagDefaultValue)

		if fieldsEnv != "" && fieldsEnv != "-" {
			if flagUsage != "" {
				flagUsage += " "
			}
			flagUsage += fmt.Sprintf("(it can also be set by env %s)", fieldsEnv)
		}

		filedType := field.Type.Kind()
		var err error
		switch filedType {

		case reflect.String:
			tempValue := ""
			flag.StringVar(&tempValue, flagKey, flagDefaultValue, flagUsage)
			if tempValue != "" {
				fieldInfo.SetString(tempValue)
			}
		case reflect.Int:
			var i int
			if flagDefaultValue != "" {
				i, err = strconv.Atoi(flagDefaultValue)
				if err != nil {
					log.Error("error parsing default to type", "error", err)
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			var tempValue int
			flag.IntVar(&tempValue, flagKey, i, flagUsage)
			if tempValue != 0 {
				fieldInfo.Set(reflect.ValueOf(tempValue))
			}
		case reflect.Int64:
			var i int64

			if flagDefaultValue != "" {
				i, err = strconv.ParseInt(flagDefaultValue, 10, 64)
				if err != nil {
					log.Error("error parsing default to type", "error", err)
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			var tempValue int64
			flag.Int64Var(&tempValue, flagKey, i, flagUsage)
			if tempValue != 0 {
				fieldInfo.SetInt(tempValue)
			}
		case reflect.Uint:
			var i uint
			if flagDefaultValue != "" {
				ii, err := strconv.ParseUint(flagDefaultValue, 10, 64)
				if err != nil {
					log.Error("error parsing default to type", "error", err)
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
				i = uint(ii)
			}
			var tempValue uint
			flag.UintVar(&tempValue, flagKey, i, flagUsage)
			if tempValue != 0 {
				fieldInfo.Set(reflect.ValueOf(tempValue))
			}
		case reflect.Uint64:
			var i uint64

			if flagDefaultValue != "" {
				i, err = strconv.ParseUint(flagDefaultValue, 10, 64)
				if err != nil {
					log.Error("error parsing default to type", "error", err)
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			var tempValue uint64
			flag.Uint64Var(&tempValue, flagKey, i, flagUsage)
			if tempValue != 0 {
				fieldInfo.SetUint(tempValue)
			}

		case reflect.Bool:
			var b bool
			if flagDefaultValue != "" {
				b, err = strconv.ParseBool(flagDefaultValue)
				if err != nil {
					log.Error("error parsing default to type", "error", err)
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			var tempValue bool
			flag.BoolVar(&tempValue, flagKey, b, flagUsage)
			if tempValue {
				fieldInfo.SetBool(tempValue)
			}
		case reflect.Struct:
			err = setUpFlags(flagKey, fieldsEnv, splitter, value)
			if err != nil {
				return err
			}

		default:
			log.Warn("unsupported type", "supported types", supportedTypes, "suggest", `"you can ignore it by flag:"-"`)
			continue
		}

		log.Debug("done process field",
			"name", field.Name,
			"type", filedType,
			"flag-key", flagKey,
			"value", value,
			"flag-usage", flagUsage,
		)
	}

	return nil
}

func SetUpEnv(splitter string, p any) error {
	return setUpEnv("", splitter, p)
}

func setUpEnv(prefix, splitter string, p any) error {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr {
		log.Error("expected a pointer to a struct", "type", p)
		return fmt.Errorf("expected a pointer to a struct, got %T", p)
	}
	// Dereference the pointer to get the struct value
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		log.Error("expected a pointer to a struct", "type", v.Kind())
		return fmt.Errorf("expected a pointer to a struct, got pointer to %v", v.Kind())
	}

	t := v.Type()

	// loop over the fields of the struct
	fields := reflect.VisibleFields(t)
	for _, field := range fields {

		fieldsEnv := field.Tag.Get("env")
		if fieldsEnv == "-" {
			continue
		}

		fieldInfo := v.FieldByIndex(field.Index)

		value := fieldInfo.Addr().Interface()

		if fieldsEnv == "" {
			fieldsEnv = field.Name
		}
		// if env is not set then use the parent env + the field env
		if prefix != "" {
			fieldsEnv = prefix + splitter + fieldsEnv
		}

		log := log.With(
			"name", field.Name,
			"type", field.Type,
			"env-key", fieldsEnv,
			"value", fmt.Sprintf("%+v", fieldInfo.Interface()),
		)

		if !fieldInfo.CanSet() {
			log.Error("cannot set field")
			return fmt.Errorf("cannot set field %v%v%s", prefix, splitter, field.Name)
		}

		fieldKind := field.Type.Kind()
		envValue, hasValue := os.LookupEnv(fieldsEnv)

		if !hasValue && fieldKind != reflect.Struct {
			continue
		}
		errorMessage := fmt.Sprintf("error parsing env to type %q value for field %q: %s", field.Type.Kind(), field.Name, envValue)
		supportedTypes := []string{"string", "int", "int8", "int16", "int32", "int64", "uint", "uint16", "uint32", "uint64", "bool", "struct"}

		switch field.Type.Kind() {
		case reflect.String:
			fieldInfo.Set(reflect.ValueOf(envValue))
		case reflect.Int:
			i, err := strconv.Atoi(envValue)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.Set(reflect.ValueOf(i))
		case reflect.Int8:
			i, _ := strconv.ParseInt(envValue, 10, 8)
			fieldInfo.Set(reflect.ValueOf(int8(i)))
		case reflect.Int16:
			i, _ := strconv.ParseInt(envValue, 10, 16)
			fieldInfo.Set(reflect.ValueOf(int16(i)))
		case reflect.Int32:
			i, err := strconv.ParseInt(envValue, 10, 32)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.Set(reflect.ValueOf(int16(i)))
		case reflect.Int64:
			i, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.SetInt(i)
		case reflect.Uint:
			i, err := strconv.ParseUint(envValue, 10, 64)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.Set(reflect.ValueOf(uint(i)))
		case reflect.Uint16:
			i, err := strconv.ParseUint(envValue, 10, 16)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.Set(reflect.ValueOf(uint16(i)))
		case reflect.Uint32:
			i, err := strconv.ParseUint(envValue, 10, 32)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.Set(reflect.ValueOf(uint32(i)))
		case reflect.Uint64:
			i, err := strconv.ParseUint(envValue, 10, 64)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.SetUint(i)
		case reflect.Bool:
			b, err := strconv.ParseBool(envValue)
			if err != nil {
				log.Error("error parsing env to type", "error", err)
				return fmt.Errorf("%s: %w", errorMessage, err)
			}
			fieldInfo.SetBool(b)
		case reflect.Struct:

			err := setUpEnv(fieldsEnv, splitter, value)

			if err != nil {
				return err
			}
		default:
			log.Warn("unsupported type", "suggest", `"you can ignore it by env:"-"`, "only supported type", supportedTypes)
			continue
		}

		log.Debug("done process field")
	}

	return nil
}

// SetUpFlagEnv set up the flag and env for the given struct
// first it will set up the flag, and then it will set up the env
func SetUpFlagEnv(splitter string, p any) error {
	err := SetUpEnv(splitter, p)
	if err != nil {
		log.Error("error setting up env", "error", err)
		return err
	}

	err = SetUpFlags(splitter, p)
	if err != nil {
		log.Error("error setting up flag", "error", err)
		return err
	}

	return err
}
