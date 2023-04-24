package validate

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/dev-mockingbird/errors"
	"github.com/dev-mockingbird/logf"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

type Rule struct {
	Must      []string
	Enum      []string
	Min       *int64
	Max       *int64
	Regexp    string
	Omitempty bool
}

type Rules map[string]Rule

func (r Rule) Validate(val reflect.Value, prev string, logger logf.Logfer, rules Rules) (empty bool, err error) {
	errOccurred := func() bool {
		if !r.Omitempty && empty {
			err = errors.Noticef("`%s` not allow empty", prev)
			return true
		}
		return false
	}
	switch val.Type().Kind() {
	case reflect.Slice, reflect.Array:
		empty = val.Len() == 0
		if !errOccurred() && !empty {
			err = validateReflectValue(val, prev, logger, rules)
		}
	case reflect.Interface:
		empty = val.IsNil()
		errOccurred()
	case reflect.Map:
		empty = len(val.MapKeys()) == 0
		if !errOccurred() && !empty {
			err = validateReflectValue(val, prev, logger, rules)
		}
	case reflect.String:
		empty = val.String() == ""
		if errOccurred() {
			return
		}
		sval := val.String()
		if r.Regexp != "" {
			var re *regexp.Regexp
			if re, err = regexp.Compile(r.Regexp); err != nil {
				logger.Logf(logf.Warn, "compile regexp for `%s` failed: %s", prev, err.Error())
				return
			}
			if !re.MatchString(sval) {
				err = errors.Noticef("`%s` cound be malformed", prev)
			}
		}
		if len(r.Enum) > 0 {
			if !funk.ContainsString(r.Enum, sval) {
				err = errors.Noticef("`%s` should be one of [%s], current value is [%s]", prev, strings.Join(r.Enum, ","), sval)
			}
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
				err = errors.Noticef("`%s` should be one of [%s], current value is [%d]", prev, strings.Join(r.Enum, ","), ival)
			}
		}
		if r.Min != nil && ival < *r.Min {
			err = errors.Noticef("`%s` should be great than equal [%d], current value is [%d]", prev, *r.Min, ival)
		}
		if r.Max != nil && ival > *r.Max {
			err = errors.Noticef("`%s` should be less than equal [%d], current value is [%d]", prev, *r.Max, ival)
		}
	case reflect.Struct:
		err = validateReflectValue(val, prev, logger, rules)
	case reflect.Ptr:
		empty = val.IsNil()
		if !errOccurred() && !empty {
			err = validateReflectValue(val, prev, logger, rules)
		}
	}
	return
}

func ParseValidateTag(tag string, rule *Rule, logger logf.Logfer) {
	rawrules := strings.Split(tag, ";")
	for _, rawrule := range rawrules {
		switch {
		case rawrule == "omitempty":
			rule.Omitempty = true
		default:
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
			case "range":
				rr := strings.Split(kv[1], ",")
				if len(rr) > 1 {
					min := cast.ToInt64(kv[0])
					rule.Min = &min
					max := cast.ToInt64(rr[1])
					rule.Max = &max
				} else {
					max := cast.ToInt64(rr[0])
					rule.Max = &max
				}
			}
		}
	}
}
