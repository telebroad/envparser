package flag_env_to_struct

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func SetUpFlags(splitter string, p any) error {
	return setUpFlags("", "", splitter, p)
}

func setUpFlags(prefix, env, splitter string, p any) error {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer to a struct, got %T", p)
	}
	// Dereference the pointer to get the struct value
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got pointer to %v", v.Kind())
	}

	t := v.Type()

	// loop over the fields of the struct
	fields := reflect.VisibleFields(t)
	for _, field := range fields {

		value := v.FieldByIndex(field.Index).Addr().Interface()

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
		for _, s := range tagData[1:] {
			k := strings.SplitN(s, ":", 2)
			if len(k) != 2 {
				return fmt.Errorf(`error: unsupported tag %q for field %q\n tag should look flag:"flag-name,default:some value,usage:help info for this flag"`, s, field.Name)
			}

			switch strings.TrimSpace(k[0]) {
			case "default":
				flagDefaultValue = k[1]
			case "usage":
				flagUsage = k[1]
			default:
				return fmt.Errorf("error: unsupported tag %q for field %q\n", s, field.Name)
			}
		}
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
			flag.StringVar(value.(*string), flagKey, flagDefaultValue, flagUsage)
		case reflect.Int:
			var i int
			if flagDefaultValue != "" {
				i, err = strconv.Atoi(flagDefaultValue)
				if err != nil {
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			flag.IntVar(value.(*int), flagKey, i, flagUsage)
		case reflect.Int64:
			var i int64

			if flagDefaultValue != "" {
				i, err = strconv.ParseInt(flagDefaultValue, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			flag.Int64Var(value.(*int64), flagKey, i, flagUsage)
		case reflect.Bool:
			var b bool
			if flagDefaultValue != "" {
				b, err = strconv.ParseBool(flagDefaultValue)
				if err != nil {
					return fmt.Errorf("error parsing default to type %q value for field %q: %s", filedType, field.Name, err)
				}
			}
			flag.BoolVar(value.(*bool), flagKey, b, flagUsage)
		case reflect.Struct:
			err = setUpFlags(flagKey, fieldsEnv, splitter, value)
			if err != nil {
				return err
			}

		default:
			fmt.Printf(`"Warning: unsupported type %q on filed %q you can igonere it by flag:"-"`, field.Type.Kind(), flagKey)
			continue
		}

		fmt.Printf("Key: %q\tType: %q\tflag-name: %q\t usage: %q\n", field.Name, filedType, flagKey, flagUsage)
	}

	return nil
}

func SetUpEnv(splitter string, p any) error {
	return setUpEnv("", splitter, p)
}
func setUpEnv(prefix, splitter string, p any) error {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer to a struct, got %T", p)
	}
	// Dereference the pointer to get the struct value
	v = v.Elem()
	if v.Kind() != reflect.Struct {
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
		if !fieldInfo.CanSet() {
			return fmt.Errorf("cannot set field %v%v%s", prefix, splitter, field.Name)
		}
		envValue, hasValue := os.LookupEnv(fieldsEnv)
		if !hasValue {
			continue
		}
		switch field.Type.Kind() {
		case reflect.String:
			fieldInfo.Set(reflect.ValueOf(envValue))
		case reflect.Int:
			i, _ := strconv.Atoi(envValue)
			fieldInfo.Set(reflect.ValueOf(i))
		case reflect.Int8:
			i, _ := strconv.ParseInt(envValue, 10, 8)
			fieldInfo.Set(reflect.ValueOf(int8(i)))
		case reflect.Int16:
			i, _ := strconv.ParseInt(envValue, 10, 16)
			fieldInfo.Set(reflect.ValueOf(int16(i)))
		case reflect.Int32:
			i, _ := strconv.ParseInt(envValue, 10, 32)
			fieldInfo.Set(reflect.ValueOf(int16(i)))
		case reflect.Int64:
			i, _ := strconv.ParseInt(envValue, 10, 64)
			fieldInfo.Set(reflect.ValueOf(i))
		case reflect.Uint:
			i, _ := strconv.ParseUint(envValue, 10, 64)
			fieldInfo.Set(reflect.ValueOf(uint(i)))
		case reflect.Uint16:
			i, _ := strconv.ParseUint(envValue, 10, 16)
			fieldInfo.Set(reflect.ValueOf(uint16(i)))
		case reflect.Uint32:
			i, _ := strconv.ParseUint(envValue, 10, 32)
			fieldInfo.Set(reflect.ValueOf(uint32(i)))
		case reflect.Uint64:
			i, _ := strconv.ParseUint(envValue, 10, 64)
			fieldInfo.Set(reflect.ValueOf(i))
		case reflect.Bool:
			b, _ := strconv.ParseBool(envValue)
			fieldInfo.Set(reflect.ValueOf(b))
		case reflect.Struct:
			err := setUpEnv(fieldsEnv, splitter, value)
			if err != nil {
				return err
			}

		default:
			continue
		}

		fmt.Printf("Key: %s\tType: %s,%s\n", field.Name, field.Type, field.Tag)
	}

	return nil
}
