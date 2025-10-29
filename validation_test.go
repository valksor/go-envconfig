package envconfig

import (
	"reflect"
	"strings"
	"testing"
)

const (
	InvalidAlphanumericValue      = "abc@123"
	EmptyStringTestCase           = "empty string"
	NonEmptyStringTestCase        = "non-empty string"
	ValidateStructExpectedError   = "ValidateStruct() expected error but got none"
	ValidateStructErrorFormat     = "ValidateStruct() error = %v, want to contain %v"
	ValidateStructUnexpectedError = "ValidateStruct() unexpected error = %v"
)

type BasicTestStruct struct {
	RequiredField string `required:"true"`
	OptionalField string
	MinField      string `min:"3"`
	MaxField      string `max:"10"`
	PatternField  string `pattern:"alphanumeric"`
	ComboField    string `required:"true" min:"2" max:"5"`
}

type NestedStruct struct {
	NestedRequired string `required:"true"`
	NestedOptional string
}

type NestedTestStruct struct {
	RequiredField string `required:"true"`
	Nested        NestedStruct
}

type PointerNestedStruct struct {
	NestedRequired string `required:"true"`
}

type PointerTestStruct struct {
	RequiredField string `required:"true"`
	NestedPtr     *PointerNestedStruct
}

type testCase struct {
	name      string
	config    any
	wantError bool
	errorMsg  string
}

type nestedTestCase struct {
	name      string
	config    NestedTestStruct
	wantError bool
	errorMsg  string
}

type pointerTestCase struct {
	name      string
	config    PointerTestStruct
	wantError bool
	errorMsg  string
}

func assertValidationResult(t *testing.T, err error, wantError bool, errorMsg string) {
	if wantError {
		if err == nil {
			t.Error(ValidateStructExpectedError)
			return
		}
		if errorMsg != "" && !strings.Contains(err.Error(), errorMsg) {
			t.Errorf(ValidateStructErrorFormat, err.Error(), errorMsg)
		}
	} else {
		if err != nil {
			t.Errorf(ValidateStructUnexpectedError, err)
		}
	}
}

func TestValidationErrorError(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "validation error for field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationErrorsError(t *testing.T) {
	tests := []struct {
		name     string
		errors   ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   ValidationErrors{},
			expected: "no validation errors",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "field1", Message: "error1"},
			},
			expected: "validation failed with 1 error(s): validation error for field 'field1': error1",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "field1", Message: "error1"},
				{Field: "field2", Message: "error2"},
			},
			expected: "validation failed with 2 error(s): validation error for field 'field1': error1; validation error for field 'field2': error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.expected {
				t.Errorf("ValidationErrors.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Error("NewValidator() returned nil")
	}

	// Check that it implements the Validator interface
	var _ Validator = validator
}

func TestStructValidatorValidateStructEdgeCases(t *testing.T) {
	validator := NewValidator()

	tests := []testCase{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
			errorMsg:  "configuration cannot be nil",
		},
		{
			name:      "nil pointer",
			config:    (*BasicTestStruct)(nil),
			wantError: true,
			errorMsg:  "configuration pointer cannot be nil",
		},
		{
			name:      "non-struct",
			config:    "not a struct",
			wantError: true,
			errorMsg:  "configuration must be a struct",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidateStructBasicValidation(t *testing.T) {
	validator := NewValidator()

	tests := []testCase{
		{
			name: "valid struct",
			config: &BasicTestStruct{
				RequiredField: "required",
				OptionalField: "optional",
				MinField:      "min",
				MaxField:      "max",
				PatternField:  "abc123",
				ComboField:    "combo",
			},
			wantError: false,
			errorMsg:  "",
		},
		{
			name: "missing required field",
			config: &BasicTestStruct{
				RequiredField: "",
				OptionalField: "optional",
				MinField:      "",
				MaxField:      "",
				PatternField:  "",
				ComboField:    "",
			},
			wantError: true,
			errorMsg:  "requiredfield",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidateStructLengthValidation(t *testing.T) {
	validator := NewValidator()

	tests := []testCase{
		{
			name: "field too short",
			config: &BasicTestStruct{
				RequiredField: "required",
				OptionalField: "",
				MinField:      "ab",
				MaxField:      "",
				PatternField:  "",
				ComboField:    "",
			},
			wantError: true,
			errorMsg:  "minimum length",
		},
		{
			name: "field too long",
			config: &BasicTestStruct{
				RequiredField: "required",
				OptionalField: "",
				MinField:      "",
				MaxField:      "this is way too long",
				PatternField:  "",
				ComboField:    "",
			},
			wantError: true,
			errorMsg:  "maximum length",
		},
		{
			name: "combo field validation",
			config: &BasicTestStruct{
				RequiredField: "required",
				OptionalField: "",
				MinField:      "",
				MaxField:      "",
				PatternField:  "",
				ComboField:    "a",
			},
			wantError: true,
			errorMsg:  "minimum length",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidateStructPatternValidation(t *testing.T) {
	validator := NewValidator()

	tests := []testCase{
		{
			name: "pattern mismatch",
			config: &BasicTestStruct{
				RequiredField: "required",
				OptionalField: "",
				MinField:      "",
				MaxField:      "",
				PatternField:  InvalidAlphanumericValue,
				ComboField:    "",
			},
			wantError: true,
			errorMsg:  "does not match required pattern",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidateNestedStructValidCases(t *testing.T) {
	validator := NewValidator()

	tests := []nestedTestCase{
		{
			name: "valid nested struct",
			config: NestedTestStruct{
				RequiredField: "required",
				Nested: NestedStruct{
					NestedRequired: "nested_required",
					NestedOptional: "nested_optional",
				},
			},
			wantError: false,
			errorMsg:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidateNestedStructErrorCases(t *testing.T) {
	validator := NewValidator()

	tests := []nestedTestCase{
		{
			name: "missing nested required field",
			config: NestedTestStruct{
				RequiredField: "required",
				Nested: NestedStruct{
					NestedRequired: "",
					NestedOptional: "nested_optional",
				},
			},
			wantError: true,
			errorMsg:  "nested.nestedrequired",
		},
		{
			name: "missing top-level required field",
			config: NestedTestStruct{
				RequiredField: "",
				Nested: NestedStruct{
					NestedRequired: "nested_required",
					NestedOptional: "",
				},
			},
			wantError: true,
			errorMsg:  "requiredfield",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidatePointerFieldValidCases(t *testing.T) {
	validator := NewValidator()

	tests := []pointerTestCase{
		{
			name: "valid pointer field",
			config: PointerTestStruct{
				RequiredField: "required",
				NestedPtr: &PointerNestedStruct{
					NestedRequired: "nested_required",
				},
			},
			wantError: false,
			errorMsg:  "",
		},
		{
			name: "nil pointer field",
			config: PointerTestStruct{
				RequiredField: "required",
				NestedPtr:     nil,
			},
			wantError: false,
			errorMsg:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorValidatePointerFieldErrorCases(t *testing.T) {
	validator := NewValidator()

	tests := []pointerTestCase{
		{
			name: "invalid nested field in pointer",
			config: PointerTestStruct{
				RequiredField: "required",
				NestedPtr: &PointerNestedStruct{
					NestedRequired: "",
				},
			},
			wantError: true,
			errorMsg:  "nestedptr.nestedrequired",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tc.config)
			assertValidationResult(t, err, tc.wantError, tc.errorMsg)
		})
	}
}

func TestStructValidatorGetFieldName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    reflect.StructField
		prefix   string
		expected string
	}{
		{
			name:     "field with mapstructure tag",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: `mapstructure:"custom_name"`, Offset: 0, Index: nil, Anonymous: false},
			prefix:   "",
			expected: "custom_name",
		},
		{
			name:     "field without mapstructure tag",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: "", Offset: 0, Index: nil, Anonymous: false},
			prefix:   "",
			expected: "testfield",
		},
		{
			name:     "field with prefix",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: "", Offset: 0, Index: nil, Anonymous: false},
			prefix:   "parent",
			expected: "parent.testfield",
		},
		{
			name:     "field with mapstructure tag and prefix",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: `mapstructure:"custom_name"`, Offset: 0, Index: nil, Anonymous: false},
			prefix:   "parent",
			expected: "parent.custom_name",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.getFieldName(testCase.field, testCase.prefix)
			if result != testCase.expected {
				t.Errorf("getFieldName() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorIsRequired(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    reflect.StructField
		expected bool
	}{
		{
			name:     "required field",
			field:    reflect.StructField{Name: "", PkgPath: "", Type: reflect.TypeOf(""), Tag: `required:"true"`, Offset: 0, Index: nil, Anonymous: false},
			expected: true,
		},
		{
			name:     "not required field",
			field:    reflect.StructField{Name: "", PkgPath: "", Type: reflect.TypeOf(""), Tag: `required:"false"`, Offset: 0, Index: nil, Anonymous: false},
			expected: false,
		},
		{
			name:     "field without required tag",
			field:    reflect.StructField{Name: "", PkgPath: "", Type: reflect.TypeOf(""), Tag: "", Offset: 0, Index: nil, Anonymous: false},
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.isRequired(testCase.field)
			if result != testCase.expected {
				t.Errorf("isRequired() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorIsEmpty(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		value    reflect.Value
		expected bool
	}{
		{
			name:     EmptyStringTestCase,
			value:    reflect.ValueOf(""),
			expected: true,
		},
		{
			name:     NonEmptyStringTestCase,
			value:    reflect.ValueOf("hello"),
			expected: false,
		},
		{
			name:     "nil pointer",
			value:    reflect.ValueOf((*string)(nil)),
			expected: true,
		},
		{
			name:     "non-nil pointer",
			value:    func() reflect.Value { s := "test"; return reflect.ValueOf(&s) }(),
			expected: false,
		},
		{
			name:     "empty slice",
			value:    reflect.ValueOf([]string{}),
			expected: true,
		},
		{
			name:     "non-empty slice",
			value:    reflect.ValueOf([]string{"item"}),
			expected: false,
		},
		{
			name:     "empty array",
			value:    reflect.ValueOf([0]string{}),
			expected: true,
		},
		{
			name:     "non-empty array",
			value:    reflect.ValueOf([1]string{"item"}),
			expected: false,
		},
		{
			name:     "empty map",
			value:    reflect.ValueOf(map[string]string{}),
			expected: true,
		},
		{
			name:     "non-empty map",
			value:    reflect.ValueOf(map[string]string{"key": "value"}),
			expected: false,
		},
		{
			name:     "struct (never empty)",
			value:    reflect.ValueOf(struct{}{}),
			expected: false,
		},
		{
			name:     "int (never empty)",
			value:    reflect.ValueOf(0),
			expected: false,
		},
		{
			name:     "bool (never empty)",
			value:    reflect.ValueOf(false),
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.isEmpty(testCase.value)
			if result != testCase.expected {
				t.Errorf("isEmpty() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorParseInt(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		input      string
		defaultVal int
		expected   int
	}{
		{
			name:       "valid integer",
			input:      "123",
			defaultVal: 0,
			expected:   123,
		},
		{
			name:       "negative integer",
			input:      "-456",
			defaultVal: 0,
			expected:   -456,
		},
		{
			name:       "invalid string",
			input:      "abc",
			defaultVal: 10,
			expected:   10,
		},
		{
			name:       EmptyStringTestCase,
			input:      "",
			defaultVal: 5,
			expected:   5,
		},
		{
			name:       "float string",
			input:      "123.45",
			defaultVal: 0,
			expected:   123, // parseInt reads up to the first non-digit
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.parseInt(testCase.input, testCase.defaultVal)
			if result != testCase.expected {
				t.Errorf("parseInt() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorMatchesPattern(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		pattern  string
		expected bool
	}{
		{
			name:     "alphanumeric valid",
			input:    "abc123",
			pattern:  "alphanumeric",
			expected: true,
		},
		{
			name:     "alphanumeric invalid",
			input:    InvalidAlphanumericValue,
			pattern:  "alphanumeric",
			expected: false,
		},
		{
			name:     "alphanumeric empty",
			input:    "",
			pattern:  "alphanumeric",
			expected: true,
		},
		{
			name:     "unknown pattern",
			input:    "anything",
			pattern:  "unknown",
			expected: true, // Unknown patterns return true
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.matchesPattern(testCase.input, testCase.pattern)
			if result != testCase.expected {
				t.Errorf("matchesPattern() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorIsAlphanumeric(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid alphanumeric",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "valid letters only",
			input:    "abcDEF",
			expected: true,
		},
		{
			name:     "valid numbers only",
			input:    "123456",
			expected: true,
		},
		{
			name:     EmptyStringTestCase,
			input:    "",
			expected: true,
		},
		{
			name:     "with special characters",
			input:    InvalidAlphanumericValue,
			expected: false,
		},
		{
			name:     "with spaces",
			input:    "abc 123",
			expected: false,
		},
		{
			name:     "with underscore",
			input:    "abc_123",
			expected: false,
		},
		{
			name:     "with dash",
			input:    "abc-123",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := validator.isAlphanumeric(testCase.input)
			if result != testCase.expected {
				t.Errorf("isAlphanumeric() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestStructValidatorValidateStructWithUnexportedFields(t *testing.T) {
	validator := NewValidator()

	type TestStruct struct {
		ExportedField   string `required:"true"`
		unexportedField string `required:"true"` // This should be ignored
	}

	// Only exported fields should be validated
	config := TestStruct{
		ExportedField:   "",
		unexportedField: "",
	}

	err := validator.ValidateStruct(&config)
	if err == nil {
		t.Error("ValidateStruct() expected error for missing exported field")
		return
	}

	// Should only complain about the exported field
	if !strings.Contains(err.Error(), "exportedfield") {
		t.Errorf("ValidateStruct() error = %v, want to contain 'exportedfield'", err.Error())
	}

	// Should not complain about unexported field
	if strings.Contains(err.Error(), "unexportedfield") {
		t.Errorf("ValidateStruct() error = %v, should not contain 'unexportedfield'", err.Error())
	}
}
