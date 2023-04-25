package validate

import (
	"testing"
)

type SimpleValidateCase struct {
	PtrIntId *int
	Str      string
}

type RegexpCase struct {
	Str string `validate:"regexp:^\\d{10}$"`
}

type EnumStrCase struct {
	Str string `validate:"enum:a,b,c"`
}

type EnumIntCase struct {
	Int int `validate:"enum:1,2,3"`
}

type MinMaxIntCase struct {
	Int int `validate:"min:1;max:10"`
}

type NestedCase struct {
	A struct {
		AA string
	}
	B *struct {
		BB *int
	}
}

type UnexportedCase struct {
	unexported bool
	Str        string `validate:"omitempty"`
}

func TestValidate_simple(t *testing.T) {
	r := SimpleValidateCase{}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.PtrIntId` not allow empty" {
			t.Fatal(err)
		}
	}
	var i int = 0
	r = SimpleValidateCase{
		PtrIntId: &i,
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Str` not allow empty" {
			t.Fatal(err)
		}
	}
}

func TestValidate_regexp(t *testing.T) {
	r := RegexpCase{}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Str` not allow empty" {
			t.Fatal(err)
		}
	}
	r = RegexpCase{
		Str: "hello",
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Str` cound be malformed" {
			t.Fatal(err)
		}
	}
	r = RegexpCase{
		Str: "1111111111",
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_enumString(t *testing.T) {
	r := EnumStrCase{}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Str` not allow empty" {
			t.Fatal(err)
		}
	}
	r = EnumStrCase{
		Str: "d",
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Str` should be one of [a,b,c], current value is [d]" {
			t.Fatal(err)
		}
	}
	r = EnumStrCase{
		Str: "a",
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_enumInt(t *testing.T) {
	r := EnumIntCase{
		Int: 4,
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Int` should be one of [1,2,3], current value is [4]" {
			t.Fatal(err)
		}
	}
	r = EnumIntCase{
		Int: 3,
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_minmax(t *testing.T) {
	r := MinMaxIntCase{
		Int: 0,
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Int` should be greater than equal [1], current value is [0]" {
			t.Fatal(err)
		}
	}
	r = MinMaxIntCase{
		Int: 11,
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.Int` should be less than equal [10], current value is [11]" {
			t.Fatal(err)
		}
	}
	r = MinMaxIntCase{
		Int: 5,
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_nested(t *testing.T) {
	r := NestedCase{}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.A.AA` not allow empty" {
			t.Fatal(err)
		}
	}
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.B` not allow empty" {
			t.Fatal(err)
		}
	}
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.B.BB` not allow empty" {
			t.Fatal(err)
		}
	}
	b := 0
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{BB: &b},
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_slice(t *testing.T) {
	r := []NestedCase{
		{},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.0.A.AA` not allow empty" {
			t.Fatal(err)
		}
	}
	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
	}}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.0.B` not allow empty" {
			t.Fatal(err)
		}
	}
	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{},
	}}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.0.B.BB` not allow empty" {
			t.Fatal(err)
		}
	}
	b := 0
	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{BB: &b},
	}}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_map(t *testing.T) {
	r := map[string]NestedCase{
		"hello": {},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.hello.A.AA` not allow empty" {
			t.Fatal(err)
		}
	}
	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
		},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.hello.B` not allow empty" {
			t.Fatal(err)
		}
	}
	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{},
		},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "[invalid-data] `.hello.B.BB` not allow empty" {
			t.Fatal(err)
		}
	}
	b := 0
	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{BB: &b},
		},
	}
	if err := GetValidator().Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_rules(t *testing.T) {
	validator := GetValidator(RR(Raw{
		".*.B.BB": "omitempty",
	}))
	r := map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{},
		},
	}
	if err := validator.Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_unexported(t *testing.T) {
	validator := GetValidator()
	r := map[string]UnexportedCase{
		"hello": {},
	}
	if err := validator.Validate(r); err != nil {
		t.Fatal(err)
	}
}
