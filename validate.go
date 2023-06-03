package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dev-mockingbird/errors"
	"github.com/dev-mockingbird/logf"
	_ "github.com/dev-mockingbird/validate/catalog"
	"github.com/ettle/strcase"
	"github.com/spf13/cast"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	OriginCase = iota // follow the definition
	SnakeCase         // hello_world
	PascalCase        // HelloWorld
	CamelCase         // helloWorld
	KebabCase         // hello-world
)

type Validator interface {
	Validate(data any, rules ...Rules) error
}

type Validate func(data any, rules ...Rules) error

func (v Validate) Validate(data any, rules ...Rules) error {
	return v(data, rules...)
}

func (v *validator) validate(data any, prev string) error {
	val := reflect.ValueOf(data)
	return v.validateReflectValue(val, prev)
}

func (v *validator) concatName(prev, n string) string {
	name := n
	switch v.nameCase {
	case CamelCase:
		name = strcase.ToCamel(n)
	case SnakeCase:
		name = strcase.ToSnake(n)
	case PascalCase:
		name = strcase.ToPascal(n)
	case KebabCase:
		name = strcase.ToKebab(n)
	}
	return fmt.Sprintf("%s.%s", prev, name)
}

func (validator *validator) getRule(name, rawrule string) (rule Rule) {
	ruleOf := func(r any) Rule {
		var ret Rule
		if raw, ok := r.(string); ok {
			ParseValidateTag(raw, &ret, validator.logger)
			return ret
		} else if ret, ok = r.(Rule); ok {
			return ret
		} else if callback, ok := r.(func(any) error); ok {
			return Rule{Callback: callback}
		}
		validator.logger.Logf(logf.Error, "can't find rule for [%s] with %#v", name, validator.rules)
		return ret
	}
	var ok bool
	var r any
	r, ok = validator.rules[name]
	defer func() {
		rule.validator = validator
		if rawrule != "" {
			ParseValidateTag(rawrule, &rule, validator.logger)
		}
	}()
	if ok {
		rule = ruleOf(r)
		return
	}
	for n, r := range validator.rules {
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

func (validator *validator) validateReflectValue(val reflect.Value, prev string) error {
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
			fn := validator.concatName(prev, f.Name)
			var rawrule string
			if tag := f.Tag.Get("validate"); tag != "" {
				rawrule = tag
			} else if tag := f.Tag.Get("json"); !validator.omitJSONTag && tag != "" {
				ts := strings.Split(tag, ",")
				rawrule = "name:" + ts[0]
				if len(ts) > 1 && ts[1] == "omitempty" {
					rawrule += ";omitempty"
				}
			}
			rule := validator.getRule(fn, rawrule)
			empty, err := rule.Validate(val.Field(i), fn)
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
			k := fmt.Sprintf("%s.%d", prev, i)
			v := val.Index(i)
			empty, err := validator.getRule(k, "").Validate(v, k)
			if err != nil {
				return err
			}
			emptyes[k] = empty
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			k := fmt.Sprintf("%s.%s", prev, cast.ToString(key.Interface()))
			v := val.MapIndex(key)
			empty, err := validator.getRule(k, "").Validate(v, k)
			if err != nil {
				return err
			}
			emptyes[k] = empty
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
			return errors.New(validator.printer.Sprintf("at least one of [%s] should be valued", strings.Join(fields, ",")), InvalidData)
		}
	}
	return nil
}

type validator struct {
	logger      logf.Logfer
	printer     *message.Printer
	rules       Rules
	nameCase    int
	omitJSONTag bool
}

type Option func(*validator)

func Logger(logger logf.Logfer) Option {
	return func(opts *validator) {
		opts.logger = logger
	}
}

func With(rules Rules) Option {
	return func(opts *validator) {
		opts.rules = rules
	}
}

func NameCase(namecase int) Option {
	return func(opts *validator) {
		opts.nameCase = namecase
	}
}

func OmitJSONTag() Option {
	return func(opts *validator) {
		opts.omitJSONTag = true
	}
}

func (validator *validator) With(rules ...Rules) *validator {
	ret := *validator
	if len(rules) > 0 {
		for _, rule := range rules {
			for k, v := range rule {
				ret.rules[k] = v
			}
		}
	}
	return &ret
}

func (v validator) Config(opt ...Option) *validator {
	ret := v
	for _, apply := range opt {
		apply(&ret)
	}
	if ret.logger == nil {
		ret.logger = logf.New()
	}
	if ret.rules == nil {
		ret.rules = make(Rules)
	}
	if ret.nameCase == 0 {
		ret.nameCase = OriginCase
	}
	if ret.printer == nil {
		ret.printer = message.NewPrinter(language.English)
	}
	return &ret
}

func Printer(printer *message.Printer) Option {
	return func(v *validator) {
		v.printer = printer
	}
}

func (v *validator) Validate(data any, rules ...Rules) error {
	if len(rules) > 0 {
		v = v.With(rules...)
	}
	return v.validate(data, "")
}

func Get(opt ...Option) Validator {
	validator := validator{}
	return validator.Config(opt...)
}

func GetValidator(opt ...Option) Validator {
	return Get(opt...)
}
