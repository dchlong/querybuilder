package field

import (
	"go/types"
	"testing"
)

// TestTimeTypePattern_DefaultConfiguration tests that default time types are properly configured
func TestTimeTypePattern_DefaultConfiguration(t *testing.T) {
	expectedDefaultTypes := map[string]bool{
		"time.Time":          true,
		"datatypes.Date":     true,
		"datatypes.Time":     true,
		"datatypes.DateTime": true,
		"sql.NullTime":       true,
		"pq.NullTime":        true,
	}

	// Check all expected defaults are present
	for expectedPattern, expectedNumeric := range expectedDefaultTypes {
		found := false
		for _, pattern := range DefaultTimeTypes {
			if pattern.Pattern == expectedPattern {
				found = true
				if pattern.IsNumeric != expectedNumeric {
					t.Errorf("Pattern %s: expected IsNumeric=%v, got %v",
						expectedPattern, expectedNumeric, pattern.IsNumeric)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected default time type pattern not found: %s", expectedPattern)
		}
	}
}

// TestInfoGenerator_NewWithDefaults tests that new generators include default time types
func TestInfoGenerator_NewWithDefaults(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	generator := NewInfoGenerator(pkg)

	if len(generator.timeTypes) == 0 {
		t.Error("NewInfoGenerator should include default time types")
	}

	// Test that time.Time is detected
	pattern := generator.matchTimeType("time.Time")
	if pattern == nil {
		t.Error("time.Time should be detected by default generator")
	} else if !pattern.IsNumeric {
		t.Error("time.Time should be marked as numeric")
	}
}

// TestInfoGenerator_CustomTimeTypes tests custom time type configuration
func TestInfoGenerator_CustomTimeTypes(t *testing.T) {
	pkg := types.NewPackage("test", "test")

	customTypes := []TimeTypePattern{
		{Pattern: "custom.DateTime", IsNumeric: true},
		{Pattern: "custom.Date", IsNumeric: false}, // Non-numeric time type
	}

	generator := NewInfoGeneratorWithTimeTypes(pkg, customTypes)

	// Test custom types are detected
	dateTimePattern := generator.matchTimeType("custom.DateTime")
	if dateTimePattern == nil {
		t.Error("custom.DateTime should be detected")
	} else if !dateTimePattern.IsNumeric {
		t.Error("custom.DateTime should be marked as numeric")
	}

	datePattern := generator.matchTimeType("custom.Date")
	if datePattern == nil {
		t.Error("custom.Date should be detected")
	} else if datePattern.IsNumeric {
		t.Error("custom.Date should not be marked as numeric")
	}

	// Test default types are not present
	if generator.matchTimeType("time.Time") != nil {
		t.Error("time.Time should not be detected in custom generator")
	}
}

// TestInfoGenerator_AddTimeType tests dynamic time type addition
func TestInfoGenerator_AddTimeType(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	generator := NewInfoGenerator(pkg)

	// Add custom time type
	generator.AddTimeType("custom.Timestamp", false)

	// Test it's detected
	pattern := generator.matchTimeType("custom.Timestamp")
	if pattern == nil {
		t.Error("Added time type should be detected")
	} else if pattern.IsNumeric {
		t.Error("Added time type should respect IsNumeric setting")
	}

	// Test original defaults still work
	if generator.matchTimeType("time.Time") == nil {
		t.Error("Default time types should still be available after adding custom types")
	}
}

// TestInfoGenerator_MatchTimeType tests the matching logic
func TestInfoGenerator_MatchTimeType(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	generator := NewInfoGenerator(pkg)

	tests := []struct {
		name            string
		typeName        string
		shouldMatch     bool
		expectedNumeric bool
	}{
		{"exact match time.Time", "time.Time", true, true},
		{"exact match datatypes.Date", "datatypes.Date", true, true},
		{"no match for similar", "time.Duration", false, false},
		{"no match for substring", "MyTime", false, false},
		{"no match for empty", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := generator.matchTimeType(tt.typeName)

			if tt.shouldMatch {
				if pattern == nil {
					t.Errorf("Expected match for %s, got nil", tt.typeName)
				} else if pattern.IsNumeric != tt.expectedNumeric {
					t.Errorf("Expected IsNumeric=%v for %s, got %v",
						tt.expectedNumeric, tt.typeName, pattern.IsNumeric)
				}
			} else {
				if pattern != nil {
					t.Errorf("Expected no match for %s, got %+v", tt.typeName, pattern)
				}
			}
		})
	}
}

// TestInfoGenerator_CreateTimeFieldInfo tests time field info creation with patterns
func TestInfoGenerator_CreateTimeFieldInfo(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	generator := NewInfoGenerator(pkg)

	baseInfo := BaseInfo{
		Name:     "CreatedAt",
		DBName:   "created_at",
		TypeName: "time.Time",
	}

	tests := []struct {
		name            string
		pattern         TimeTypePattern
		expectedNumeric bool
	}{
		{
			name:            "numeric time type",
			pattern:         TimeTypePattern{Pattern: "time.Time", IsNumeric: true},
			expectedNumeric: true,
		},
		{
			name:            "non-numeric time type",
			pattern:         TimeTypePattern{Pattern: "custom.Date", IsNumeric: false},
			expectedNumeric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := generator.createTimeFieldInfo(baseInfo, tt.pattern)

			if info == nil {
				t.Fatal("createTimeFieldInfo returned nil")
			}

			if !info.IsTime {
				t.Error("Time field should have IsTime=true")
			}

			if info.IsNumeric != tt.expectedNumeric {
				t.Errorf("Expected IsNumeric=%v, got %v", tt.expectedNumeric, info.IsNumeric)
			}

			if info.Name != baseInfo.Name {
				t.Errorf("Expected Name=%s, got %s", baseInfo.Name, info.Name)
			}
		})
	}
}

// TestInfoGenerator_GenFieldInfo_TimeTypes tests end-to-end time type detection
func TestInfoGenerator_GenFieldInfo_TimeTypes(t *testing.T) {
	pkg := types.NewPackage("test", "test")
	generator := NewInfoGenerator(pkg)

	tests := []struct {
		name            string
		typeName        string
		shouldDetect    bool
		expectedNumeric bool
	}{
		{"time.Time", "time.Time", true, true},
		{"datatypes.Date", "datatypes.Date", true, true},
		{"regular string", "string", false, false},
		{"time.Duration", "time.Duration", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Manually set the type name in baseInfo for testing
			baseInfo := BaseInfo{
				Name:     "TestField",
				DBName:   "test_field",
				TypeName: tt.typeName,
			}

			if tt.shouldDetect {
				pattern := generator.matchTimeType(tt.typeName)
				if pattern == nil {
					t.Errorf("Expected to detect time type %s", tt.typeName)
					return
				}

				info := generator.createTimeFieldInfo(baseInfo, *pattern)
				if !info.IsTime {
					t.Error("Detected time type should have IsTime=true")
				}
				if info.IsNumeric != tt.expectedNumeric {
					t.Errorf("Expected IsNumeric=%v for %s, got %v",
						tt.expectedNumeric, tt.typeName, info.IsNumeric)
				}
			} else {
				pattern := generator.matchTimeType(tt.typeName)
				if pattern != nil {
					t.Errorf("Should not detect %s as time type", tt.typeName)
				}
			}
		})
	}
}
