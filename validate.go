package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dev-mockingbird/errors"
	"github.com/dev-mockingbird/logf"
	"github.com/spf13/cast"
)

type Validator interface {
	Validate(data any, rules ...Rules) error
}

type Validate func(data any, rules ...Rules) error

func (v Validate) Validate(data any, rules ...Rules) error {
	return v(data, rules...)
}

func validate(data any, prev string, logger logf.Logfer, omitJsonTag bool, rules Rules) error {
	val := reflect.ValueOf(data)
	return validateReflectValue(val, prev, logger, omitJsonTag, rules)
}

func getRule(name string, rules Rules, rawrule string, logger logf.Logfer) (rule Rule) {
	ruleOf := func(r any) Rule {
		var ret Rule
		if raw, ok := r.(string); ok {
			ParseValidateTag(raw, &ret, logger)
			return ret
		} else if ret, ok = r.(Rule); ok {
			return ret
		} else if callback, ok := r.(func(any) error); ok {
			return Rule{Callback: callback}
		}
		logger.Logf(logf.Error, "can't find rule for [%s] with %#v", name, rules)
		return ret
	}
	var ok bool
	var r any
	r, ok = rules[name]
	defer func() {
		if rawrule != "" {
			ParseValidateTag(rawrule, &rule, logger)
		}
	}()
	if ok {
		rule = ruleOf(r)
		return
	}
	for n, r := range rules {
		ns := strings.Split(n, ".")
		names := strings.Split(name, ".")
		if len(ns) != len(names) {
			break
		}
		var notfound bool
		for i := 0; i < len(ns); i++ {
			if ns[i] == "*" || ns[i] == names[i] {
				continue
			}
			notfound = true
			break
		}
		if !notfound {
			rule = ruleOf(r)
			break
		}
	}
	return rule
}

func validateReflectValue(val reflect.Value, prev string, logger logf.Logfer, omitJsonTag bool, rules Rules) error {
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	must := make(map[string][]string)
	emptyes := make(map[string]bool)
	switch val.Type().Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Type().Field(i)
			if !f.IsExported() {
				continue
			}
			fn := fmt.Sprintf("%s.%s", prev, f.Name)
			var rawrule string
			if tag := f.Tag.Get("validate"); tag != "" {
				rawrule = tag
			} else if tag := f.Tag.Get("json"); !omitJsonTag && strings.Contains(tag, "omitempty") {
				rawrule = "omitempty"
			}
			rule := getRule(fn, rules, rawrule, logger)
			empty, err := rule.Validate(val.Field(i), fn, logger, omitJsonTag, rules)
			if err != nil {
				return err
			}
			for _, k := range rule.Must {
				must[k] = append(must[k], fn)
			}
			emptyes[fn] = empty
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			if err := validateReflectValue(val.Index(i), fmt.Sprintf("%s.%d", prev, i), logger, omitJsonTag, rules); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if err := validateReflectValue(val.MapIndex(key), fmt.Sprintf("%s.%s", prev, cast.ToString(key.Interface())), logger, omitJsonTag, rules); err != nil {
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
			return errors.Tag(fmt.Errorf("at least one of [%s] should be valued", strings.Join(fields, ",")), InvalidData)
		}
	}
	return nil
}

type validateOptions struct {
	logger      logf.Logfer
	rules       Rules
	omitJSONTag bool
}

type Option func(*validateOptions)

func Logger(logger logf.Logfer) Option {
	return func(opts *validateOptions) {
		opts.logger = logger
	}
}

func R(rules Rules) Option {
	return func(opts *validateOptions) {
		opts.rules = rules
	}
}

func OmitJSONTag() Option {
	return func(opts *validateOptions) {
		opts.omitJSONTag = true
	}
}

func applyOption(opts *validateOptions, opt ...Option) {
	for _, apply := range opt {
		apply(opts)
	}
	if opts.logger == nil {
		opts.logger = logf.New()
	}
	if opts.rules == nil {
		opts.rules = make(Rules)
	}
}

func Get(opt ...Option) Validator {
	var opts validateOptions
	applyOption(&opts, opt...)
	return Validate(func(data any, rules ...Rules) error {
		r := opts.rules
		if len(rules) > 0 {
			r = make(Rules)
			for k, v := range opts.rules {
				r[k] = v
			}
			for _, rule := range rules {
				for k, v := range rule {
					r[k] = v
				}
			}
		}
		return validate(data, "", opts.logger, opts.omitJSONTag, r)
	})
}

func GetValidator(opt ...Option) Validator {
	return Get(opt...)
}
