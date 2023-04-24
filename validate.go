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
	Validate(data any) error
}

type Validate func(data any) error

func (v Validate) Validate(data any) error {
	return v(data)
}

func validate(data any, prev string, logger logf.Logfer, rules Rules) error {
	val := reflect.ValueOf(data)
	return validateReflectValue(val, prev, logger, rules)
}

func getRule(name string, rules Rules, rawrule string, logger logf.Logfer) (rule Rule) {
	var ok bool
	rule, ok = rules[name]
	defer func() {
		if rawrule != "" {
			ParseValidateTag(rawrule, &rule, logger)
		}
	}()
	if ok {
		return rule
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
			rule = r
			break
		}
	}
	return rule
}

func validateReflectValue(val reflect.Value, prev string, logger logf.Logfer, rules Rules) error {
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
			} else if tag := f.Tag.Get("json"); tag != "" {
				if strings.Contains(tag, "omitempty") {
					rawrule = "omitempty"
				}
			}
			rule := getRule(fn, rules, rawrule, logger)
			empty, err := rule.Validate(val.Field(i), fn, logger, rules)
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
			if err := validateReflectValue(val.Index(i), fmt.Sprintf("%s.%d", prev, i), logger, rules); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			if err := validateReflectValue(val.MapIndex(key), fmt.Sprintf("%s.%s", prev, cast.ToString(key.Interface())), logger, rules); err != nil {
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

type validateOptions struct {
	logger  logf.Logfer
	rules   Rules
	rawRule Raw
}

type Option func(*validateOptions)

func Logger(logger logf.Logfer) Option {
	return func(opts *validateOptions) {
		opts.logger = logger
	}
}

type Raw map[string]string

func R(rules Rules) Option {
	return func(opts *validateOptions) {
		opts.rules = rules
	}
}

func RR(rr Raw) Option {
	return func(opts *validateOptions) {
		opts.rawRule = rr
	}
}

func GetValidator(opt ...Option) Validator {
	return Validate(func(data any) error {
		var opts validateOptions
		for _, apply := range opt {
			apply(&opts)
		}
		if opts.logger == nil {
			opts.logger = logf.New()
		}
		if opts.rules == nil {
			opts.rules = make(Rules)
		}
		for n, r := range opts.rawRule {
			var ru Rule
			ParseValidateTag(r, &ru, opts.logger)
			opts.rules[n] = ru
		}
		return validate(data, "", opts.logger, opts.rules)
	})
}
