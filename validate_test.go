package validate

import (
	"errors"
	"testing"

	"github.com/tj/assert"
)

type SimpleValidateCase struct {
	PtrIntId *int
	Str      string
}

type SignOption struct {
	Name        string `json:"name" validate:"must:id;omitempty;regexp:^[a-zA-Z0-9_\\-]*$"`
	EmailAddr   string `json:"email_addr" validate:"must:id;omitempty;regexp:^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"`
	PhoneNumber string `json:"phone_number" validate:"must:id;omitempty;regexp:^[\\+]?[(]?[0-9]{3}[)]?[-\\s\\.]?[0-9]{3}[-\\s\\.]?[0-9]{4,6}$"`
	Password    string `json:"password" validate:"must:pass;omitempty"`
	VerifyCode  string `json:"verify_code" validate:"must:pass;omitempty"`
	RemoteIP    string `json:"remote_ip"`
	Device      string `json:"device,omitempty"`
	DeviceId    string `json:"device_id,omitempty"`
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

func TestValidate_complex(t *testing.T) {
	s := SignOption{Name: "hello", Password: "1234565", RemoteIP: "127.0.0.1:6095"}
	err := Get().Validate(s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidate_simple(t *testing.T) {
	r := SimpleValidateCase{}
	err := GetValidator().Validate(r)
	assert.True(t, len(err) == 2)
	assert.Equal(t, "not allow empty", err[0].Message)
	assert.Equal(t, []string{".PtrIntId"}, err[0].Fields)
	assert.Equal(t, "not allow empty", err[1].Message)
	assert.Equal(t, []string{".Str"}, err[1].Fields)

	var i int = 0
	r = SimpleValidateCase{
		PtrIntId: &i,
	}
	err = GetValidator().Validate(r)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "not allow empty", err[0].Message)
	assert.Equal(t, []string{".Str"}, err[0].Fields)
}

func TestValidate_regexp(t *testing.T) {
	r := RegexpCase{}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.Str` not allow empty" {
			t.Fatal(err)
		}
	}
	r = RegexpCase{
		Str: "hello",
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.Str` cound be malformed" {
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
		if err.Error() != "`.Str` not allow empty" {
			t.Fatal(err)
		}
	}
	r = EnumStrCase{
		Str: "d",
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.Str` should be one of [a,b,c], current value is [d]" {
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
		if err.Error() != "`.Int` should be one of [1,2,3], current value is [4]" {
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
		if err.Error() != "`.Int` should be greater than equal [1], current value is [0]" {
			t.Fatal(err)
		}
	}
	r = MinMaxIntCase{
		Int: 11,
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.Int` should be less than equal [10], current value is [11]" {
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
	err := GetValidator().Validate(r)
	assert.True(t, len(err) == 2)
	assert.Equal(t, "not allow empty", err[0].Message)
	assert.Equal(t, []string{".A.AA"}, err[0].Fields)
	assert.Equal(t, "not allow empty", err[1].Message)
	assert.Equal(t, []string{".B"}, err[1].Fields)
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.B` not allow empty" {
			t.Fatal(err)
		}
	}
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{},
	}
	if err := GetValidator().Validate(r); err != nil {
		if err.Error() != "`.B.BB` not allow empty" {
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

func TestValidate_snakecase(t *testing.T) {
	r := NestedCase{}
	err := GetValidator(NameCase(SnakeCase)).Validate(r)
	assert.True(t, len(err) == 2)
	assert.Equal(t, "not allow empty", err[0].Message)
	assert.Equal(t, []string{".a.aa"}, err[0].Fields)
	assert.Equal(t, "not allow empty", err[1].Message)
	assert.Equal(t, []string{".b"}, err[1].Fields)
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
	}
	if err := GetValidator(NameCase(SnakeCase)).Validate(r); err != nil {
		if err.Error() != "`.b` not allow empty" {
			t.Fatal(err)
		}
	}
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{},
	}
	if err := GetValidator(NameCase(SnakeCase)).Validate(r); err != nil {
		if err.Error() != "`.b.bb` not allow empty" {
			t.Fatal(err)
		}
	}
	b := 0
	r = NestedCase{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{BB: &b},
	}
	if err := GetValidator(NameCase(SnakeCase)).Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_slice(t *testing.T) {
	r := []NestedCase{
		{},
	}
	err := GetValidator().Validate(r)
	t.Logf("err: %s", err.Error())
	assert.True(t, len(err) == 2)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".0.A.AA"}, err[0].Fields, "fileds should be .0.A.AA")
	assert.Equal(t, "not allow empty", err[1].Message, "message should not allow empty")
	assert.Equal(t, []string{".0.B"}, err[1].Fields, "fileds should be .0.B")

	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
	}}
	err = GetValidator().Validate(r)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".0.B"}, err[0].Fields, "fileds should be .0.B")

	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{},
	}}
	err = GetValidator().Validate(r)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".0.B.BB"}, err[0].Fields, "fileds should be .0.B.BB")

	b := 0
	r = []NestedCase{{
		A: struct{ AA string }{AA: "a"},
		B: &struct{ BB *int }{BB: &b},
	}}
	err = GetValidator().Validate(r)
	assert.Nil(t, err, "err should be nil")
}

func TestValidate_map(t *testing.T) {
	r := map[string]NestedCase{
		"hello": {},
	}
	err := GetValidator().Validate(r)
	t.Logf("err: %s", err.Error())
	assert.True(t, len(err) == 2)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".hello.A.AA"}, err[0].Fields, "fileds should be .hello.A.AA")
	assert.Equal(t, "not allow empty", err[1].Message, "message should not allow empty")
	assert.Equal(t, []string{".hello.B"}, err[1].Fields, "fileds should be .hello.A.AA")

	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
		},
	}
	err = GetValidator().Validate(r)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".hello.B"}, err[0].Fields, "fileds should be .hello.B")

	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{},
		},
	}
	err = GetValidator().Validate(r)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "not allow empty", err[0].Message, "message should not allow empty")
	assert.Equal(t, []string{".hello.B.BB"}, err[0].Fields, "fileds should be .hello.B.BB")
	b := 0
	r = map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{BB: &b},
		},
	}
	err = GetValidator().Validate(r)
	assert.Nil(t, err, "err should be nil")
}

func TestValidate_rules(t *testing.T) {
	validator := GetValidator(With(Rules{
		".*.B.BB": "omitempty",
	}))
	r := map[string]NestedCase{
		"hello": {
			A: struct{ AA string }{AA: "a"},
			B: &struct{ BB *int }{},
		},
	}
	err := validator.Validate(r)
	assert.Nil(t, err, "err should be nil")
}

func TestValidate_unexported(t *testing.T) {
	validator := GetValidator()
	r := map[string]UnexportedCase{
		"hello": {},
	}
	err := validator.Validate(r)
	assert.Nil(t, err, "err should be nil")
}

func TestValidate_outerrule(t *testing.T) {
	validator := GetValidator()
	r := NestedCase{}
	rules := Rules{
		".A.AA": "omitempty",
		".B":    func(val any) error { return errors.New("hello world") },
	}
	err := validator.Validate(r, rules, rules)
	assert.True(t, len(err) == 1)
	assert.Equal(t, "hello world", err[0].Message, "message should be hello world")
	assert.Equal(t, []string{".B"}, err[0].Fields, "fileds should be .B")
}

func TestValidate_IsA(t *testing.T) {
	validator := Get()
	err := validator.Validate(map[string]string{
		"password": "a123456789",
	}, Rules{
		".password": "is:password",
	})
	assert.Nil(t, err, "err should be nil")
	err = validator.Validate(map[string]string{
		"password": "123456789",
	}, Rules{
		".password": "is:password",
	})
	assert.True(t, len(err) == 1)
	assert.Equal(t, "is not one of the [password]", err[0].Message, "message not correct")
	assert.Equal(t, []string{".password"}, err[0].Fields, "fields should be .password")
}
