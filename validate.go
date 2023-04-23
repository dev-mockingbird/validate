package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/dev-mockingbird/errors"
	"github.com/dev-mockingbird/logf"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

type Validator interface {
	Validate(data any) error
}

type Validate func(data any) error

func (v Validate) Validate(data any) error {
	return v(data)
}

type validation struct {
	must      []string
	enum      []string
	min       *int64
	max       *int64
	regexp    string
	omitempty bool
}

func parseValidateTag(tag string) validation {
	var ret validation
	rules := strings.Split(tag, ";")
	for _, rule := range rules {
		switch {
		case rule == "omitempty":
			ret.omitempty = true
		default:
			kv := strings.Split(rule, ":")
			if len(kv) < 2 {
				// not supported bool option, ignore
				continue
			}
			switch kv[0] {
			case "must":
				ret.must = strings.Split(kv[1], ",")
			case "regexp":
				ret.regexp = kv[1]
			case "enum":
				ret.enum = strings.Split(kv[1], ",")
			case "min":
				min := cast.ToInt64(kv[1])
				ret.min = &min
			case "max":
				max := cast.ToInt64(kv[1])
				ret.max = &max
			case "range":
				rr := strings.Split(kv[1], ",")
				if len(rr) > 1 {
					min := cast.ToInt64(kv[0])
					ret.min = &min
					max := cast.ToInt64(rr[1])
					ret.max = &max
				} else {
					max := cast.ToInt64(rr[0])
					ret.max = &max
				}
			}
		}
	}
	return ret
}

func (v validation) validate(val reflect.Value, prev string, logger logf.Logger) (empty bool, err error) {
	errOccurred := func() bool {
		if !v.omitempty && empty {
			err = errors.Noticef("`%s` not allow empty", prev)
			return true
		}
		return false
	}
	switch val.Type().Kind() {
	case reflect.Slice, reflect.Array:
		empty = val.Len() == 0
		if !errOccurred() && !empty {
			err = validate(val.Interface(), prev, logger)
		}
	case reflect.Interface:
		empty = val.IsNil()
		errOccurred()
	case reflect.Map:
		empty = len(val.MapKeys()) == 0
		if !errOccurred() && !empty {
			err = validate(val.Interface(), prev, logger)
		}
	case reflect.String:
		empty = val.String() == ""
		if errOccurred() {
			return
		}
		sval := val.String()
		if v.regexp != "" {
			var re *regexp.Regexp
			if re, err = regexp.Compile(v.regexp); err != nil {
				logger.Logf(logf.Warn, "compile regexp for `%s` failed: %s", prev, err.Error())
				return
			}
			if !re.MatchString(sval) {
				err = errors.Noticef("`%s` cound be malformed", prev)
			}
		}
		if len(v.enum) > 0 {
			if !funk.ContainsString(v.enum, sval) {
				err = errors.Noticef("`%s` should be one of [%s], current value is [%s]", prev, strings.Join(v.enum, ","), sval)
			}
		}
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ival := val.Int()
		if len(v.enum) > 0 {
			if !funk.ContainsInt64(func() []int64 {
				ret := make([]int64, len(v.enum))
				for i, e := range v.enum {
					ret[i] = cast.ToInt64(e)
				}
				return ret
			}(), ival) {
				err = errors.Noticef("`%s` should be one of [%s], current value is [%d]", prev, strings.Join(v.enum, ","), ival)
			}
		}
		if v.min != nil && ival < *v.min {
			err = errors.Noticef("`%s` should be great than equal [%d], current value is [%d]", prev, *v.min, ival)
		}
		if v.max != nil && ival > *v.max {
			err = errors.Noticef("`%s` should be less than equal [%d], current value is [%d]", prev, *v.max, ival)
		}
	case reflect.Struct:
		err = validate(val.Interface(), prev, logger)
	case reflect.Ptr:
		empty = val.IsNil()
		if !errOccurred() && !empty {
			err = validate(val.Interface(), prev, logger)
		}
	}
	return
}

func validate(data any, prev string, logger logf.Logger) error {
	val := reflect.ValueOf(data)
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	must := make(map[string][]string)
	emptyes := make(map[string]bool)
	switch val.Type().Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Type().Field(i)
			fn := fmt.Sprintf("%s.%s", prev, f.Name)
			var v validation
			if tag := f.Tag.Get("validate"); tag != "" {
				v = parseValidateTag(tag)
			} else if tag := f.Tag.Get("json"); tag != "" {
				v.omitempty = strings.Contains(tag, "omitempty")
			}
			empty, err := v.validate(val.Field(i), fn, logger)
			if err != nil {
				return err
			}
			for _, k := range v.must {
				must[k] = append(must[k], fn)
			}
			emptyes[fn] = empty
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			if err := validate(val.Index(i), fmt.Sprintf("%s.%d", prev, i), logger); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if err := validate(val.MapIndex(key).Interface(), fmt.Sprintf("%s.%s", prev, cast.ToString(key.Interface())), logger); err != nil {
				return err
			}
		}
	}
	for _, fields := range must {
		found := false
		for _, field := range fields {
			if !emptyes[field] {
				found = true
			}
		}
		if !found {
			return errors.Noticef("at least one of [%s] should be valued", strings.Join(fields, ","))
		}
	}
	return nil
}

func GetValidator(logger logf.Logger) Validator {
	return Validate(func(data any) error {
		return validate(data, "", logger)
	})
}
