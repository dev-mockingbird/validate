package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/dev-mockingbird/logf"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

type ValidateError struct {
	Fields  []string `json:"fields"`
	Message string   `json:"message"`
}

func (v ValidateError) Error() string {
	return fmt.Sprintf("`%s` %s", strings.Join(v.Fields, ","), v.Message)
}

var _ error = &ValidateError{}

type Rule struct {
	IsA       []string
	Must      []string
	Enum      []string
	Min       *int64
	Max       *int64
	Regexp    string
	Callback  func(interface{}) error
	Omitempty bool
	validator *validator
}

type Rules map[string]any

func (r Rule) Validate(val reflect.Value, prev string) (empty bool, err *ValidateError) {
	assertEmpty := func() bool {
		if !r.Omitempty && empty {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("not allow empty"),
			}
			return true
		}
		return false
	}
	if r.Callback != nil {
		er := r.Callback(val.Interface())
		if er != nil {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: er.Error(),
			}
		}
		return
	}
	switch val.Type().Kind() {
	case reflect.Slice, reflect.Array:
		empty = val.Len() == 0
		if !assertEmpty() && !empty {
			err = r.validator.validateReflectValue(val, prev)
		}
	case reflect.Interface:
		empty = val.IsNil()
		assertEmpty()
	case reflect.Map:
		empty = len(val.MapKeys()) == 0
		if !assertEmpty() && !empty {
			err = r.validator.validateReflectValue(val, prev)
		}
	case reflect.String:
		sval := val.String()
		empty = sval == ""
		if assertEmpty() || empty {
			return
		}
		if len(r.IsA) > 0 {
			var found bool
			for _, a := range r.IsA {
				if v, ok := atoms[a]; ok {
					found = true
					if v(sval) {
						return
					}
				}
			}
			if found {
				err = &ValidateError{
					Fields:  []string{prev},
					Message: r.validator.printer.Sprintf("is not one of the [%s]", strings.Join(r.IsA, ",")),
				}
				return
			}
			r.validator.logger.Logf(logf.Warn, "not found [is a] definition for [%s]", r.IsA)
		}
		if r.Regexp != "" {
			if re, e := regexp.Compile(r.Regexp); e != nil {
				r.validator.logger.Logf(logf.Warn, "compile regexp for `%s` failed: %s", prev, err.Error())
				err = &ValidateError{Fields: []string{prev}, Message: r.validator.printer.Sprintf("can't compile regexp: %s", err.Error())}
			} else if !re.MatchString(sval) {
				err = &ValidateError{
					Fields:  []string{prev},
					Message: r.validator.printer.Sprintf("cound be malformed"),
				}
				return
			}
		}
		if len(r.Enum) > 0 && !funk.ContainsString(r.Enum, sval) {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("should be one of [%s], current value is [%s]", strings.Join(r.Enum, ","), sval),
			}
			return
		} else if r.Min != nil && len(sval) < int(*r.Min) {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("has a minimum length [%d]", *r.Min),
			}
			return
		} else if r.Max != nil && len(sval) > int(*r.Max) {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("has a maximum length [%d]", *r.Max),
			}
			return
		}
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ival := val.Int()
		if len(r.Enum) > 0 {
			if !funk.ContainsInt64(func() []int64 {
				ret := make([]int64, len(r.Enum))
				for i, e := range r.Enum {
					ret[i] = cast.ToInt64(e)
				}
				return ret
			}(), ival) {
				err = &ValidateError{
					Fields:  []string{prev},
					Message: r.validator.printer.Sprintf("should be one of [%s], current value is [%d]", strings.Join(r.Enum, ","), ival),
				}
				return
			}
		} else if r.Min != nil && ival < *r.Min {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("should be greater than equal [%d], current value is [%d]", *r.Min, ival),
			}
			return
		} else if r.Max != nil && ival > *r.Max {
			err = &ValidateError{
				Fields:  []string{prev},
				Message: r.validator.printer.Sprintf("should be less than equal [%d], current value is [%d]", *r.Max, ival),
			}
			return
		}
	case reflect.Struct:
		err = r.validator.validateReflectValue(val, prev)
	case reflect.Ptr:
		empty = val.IsNil()
		if !assertEmpty() && !empty {
			err = r.validator.validateReflectValue(val, prev)
		}
	}
	return
}

func ParseValidateTag(rawrule string, rule *Rule, logger logf.Logfer) {
	rawrules := strings.Split(rawrule, ";")
	for _, rawrule := range rawrules {
		if rawrule == "omitempty" {
			rule.Omitempty = true
			continue
		}
		kv := strings.Split(rawrule, ":")
		if len(kv) < 2 {
			logger.Logf(logf.Warn, "can't recognize rule [%s]", rawrule)
			continue
		}
		switch kv[0] {
		case "must":
			rule.Must = strings.Split(kv[1], ",")
		case "regexp":
			rule.Regexp = kv[1]
		case "enum":
			rule.Enum = strings.Split(kv[1], ",")
		case "min":
			min := cast.ToInt64(kv[1])
			rule.Min = &min
		case "max":
			max := cast.ToInt64(kv[1])
			rule.Max = &max
		case "is":
			if kv[1] == "" {
				continue
			}
			rule.IsA = strings.Split(kv[1], ",")
		case "range":
			rr := strings.Split(kv[1], ",")
			if len(rr) == 0 {
				max := cast.ToInt64(rr[0])
				rule.Max = &max
				continue
			}
			min := cast.ToInt64(kv[0])
			rule.Min = &min
			max := cast.ToInt64(rr[1])
			rule.Max = &max
		}
	}
}
