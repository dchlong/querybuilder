package field

import (
	"go/types"
	"testing"
)

// TestTimeTypeDetection_Integration tests time type detection in realistic scenarios
func TestTimeTypeDetection_Integration(t *testing.T) {
	pkg := types.NewPackage("models", "models")
	generator := NewInfoGenerator(pkg)

	tests := []struct {
		name             string
		fieldName        string
		typeName         string
		expectedDetected bool
		expectedNumeric  bool
		description      string
	}{
		{
			name:             "standard_time",
			fieldName:        "CreatedAt",
			typeName:         "time.Time",
			expectedDetected: true,
			expectedNumeric:  true,
			description:      "Standard Go time.Time should be detected as numeric time type",
		},
		{
			name:             "gorm_datatypes_date",
			fieldName:        "BirthDate",
			typeName:         "datatypes.Date",
			expectedDetected: true,
			expectedNumeric:  true,
			description:      "GORM datatypes.Date should be detected as numeric time type",
		},
		{
			name:             "sql_null_time",
			fieldName:        "DeletedAt",
			typeName:         "sql.NullTime",
			expectedDetected: true,
			expectedNumeric:  true,
			description:      "SQL NullTime should be detected as numeric time type",
		},
		{
			name:             "postgres_null_time",
			fieldName:        "ArchivedAt",
			typeName:         "pq.NullTime",
			expectedDetected: true,
			expectedNumeric:  true,
			description:      "PostgreSQL pq.NullTime should be detected as numeric time type",
		},
		{
			name:             "custom_non_time",
			fieldName:        "Name",
			typeName:         "string",
			expectedDetected: false,
			expectedNumeric:  false,
			description:      "Regular string should not be detected as time type",
		},
		{
			name:             "duration_not_time",
			fieldName:        "Timeout",
			typeName:         "time.Duration",
			expectedDetected: false,
			expectedNumeric:  false,
			description:      "time.Duration should not be detected as time type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the time type detection directly
			pattern := generator.matchTimeType(tt.typeName)
			if tt.expectedDetected {
				if pattern == nil {
					t.Errorf("%s: Expected to detect time type %s, but didn't",
						tt.description, tt.typeName)
					return
				}
				if pattern.IsNumeric != tt.expectedNumeric {
					t.Errorf("%s: Expected IsNumeric=%v for %s, got %v",
						tt.description, tt.expectedNumeric, tt.typeName, pattern.IsNumeric)
				}
			} else {
				if pattern != nil {
					t.Errorf("%s: Should not detect %s as time type, but did: %+v",
						tt.description, tt.typeName, pattern)
				}
			}

			// Test base info creation (this mimics the real field processing)
			baseInfo := BaseInfo{
				Name:     tt.fieldName,
				DBName:   tt.fieldName,
				TypeName: tt.typeName,
			}

			if tt.expectedDetected {
				timeInfo := generator.createTimeFieldInfo(baseInfo, *pattern)
				if !timeInfo.IsTime {
					t.Errorf("%s: Time field should have IsTime=true", tt.description)
				}
				if timeInfo.IsNumeric != tt.expectedNumeric {
					t.Errorf("%s: Time field should have IsNumeric=%v, got %v",
						tt.description, tt.expectedNumeric, timeInfo.IsNumeric)
				}
			}
		})
	}
}

// TestCustomTimeTypeConfiguration tests adding custom time types
func TestCustomTimeTypeConfiguration(t *testing.T) {
	pkg := types.NewPackage("models", "models")

	// Test with empty custom types (should still have defaults)
	generator1 := NewInfoGenerator(pkg)
	if len(generator1.timeTypes) == 0 {
		t.Error("Generator with defaults should have time types configured")
	}

	// Test with custom types only
	customTypes := []TimeTypePattern{
		{Pattern: "custom.LocalDateTime", IsNumeric: true},
		{Pattern: "custom.DateOnly", IsNumeric: false},
	}
	generator2 := NewInfoGeneratorWithTimeTypes(pkg, customTypes)

	if len(generator2.timeTypes) != 2 {
		t.Errorf("Custom generator should have exactly 2 time types, got %d",
			len(generator2.timeTypes))
	}

	// Test custom types work
	if generator2.matchTimeType("custom.LocalDateTime") == nil {
		t.Error("Custom LocalDateTime should be detected")
	}
	if generator2.matchTimeType("time.Time") != nil {
		t.Error("Default time.Time should not be detected in custom generator")
	}

	// Test adding time types dynamically
	generator3 := NewInfoGenerator(pkg)
	originalCount := len(generator3.timeTypes)

	generator3.AddTimeType("dynamic.Time", true)
	if len(generator3.timeTypes) != originalCount+1 {
		t.Error("AddTimeType should increase the time types count")
	}

	if generator3.matchTimeType("dynamic.Time") == nil {
		t.Error("Dynamically added time type should be detected")
	}
}

// TestTimeTypeBackwardCompatibility ensures existing behavior is preserved
func TestTimeTypeBackwardCompatibility(t *testing.T) {
	pkg := types.NewPackage("models", "models")
	generator := NewInfoGenerator(pkg)

	// These are the critical time types that must continue working
	criticalTimeTypes := []string{
		"time.Time",
		"datatypes.Date",
	}

	for _, timeType := range criticalTimeTypes {
		t.Run("backward_compatibility_"+timeType, func(t *testing.T) {
			pattern := generator.matchTimeType(timeType)
			if pattern == nil {
				t.Errorf("Critical time type %s must be detected for backward compatibility",
					timeType)
				return
			}

			if !pattern.IsNumeric {
				t.Errorf("Critical time type %s must be numeric for backward compatibility",
					timeType)
			}

			// Test field info creation
			baseInfo := BaseInfo{
				Name:     "TestField",
				DBName:   "test_field",
				TypeName: timeType,
			}

			info := generator.createTimeFieldInfo(baseInfo, *pattern)
			if !info.IsTime {
				t.Errorf("Time field info for %s should have IsTime=true", timeType)
			}
			if !info.IsNumeric {
				t.Errorf("Time field info for %s should have IsNumeric=true", timeType)
			}
		})
	}
}
