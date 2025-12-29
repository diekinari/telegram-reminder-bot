package domain

import "testing"

func TestFrequency_String(t *testing.T) {
	tests := []struct {
		freq Frequency
		want string
	}{
		{FrequencyDaily, "daily"},
		{FrequencyEveryOtherDay, "every_other_day"},
		{FrequencyWeekly, "weekly"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.freq.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrequency_DisplayName(t *testing.T) {
	tests := []struct {
		freq Frequency
		want string
	}{
		{FrequencyDaily, "Ежедневно"},
		{FrequencyEveryOtherDay, "Через день"},
		{FrequencyWeekly, "Раз в неделю"},
		{Frequency("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.freq), func(t *testing.T) {
			if got := tt.freq.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFrequency(t *testing.T) {
	tests := []struct {
		input   string
		want    Frequency
		wantOk  bool
	}{
		{"daily", FrequencyDaily, true},
		{"every_other_day", FrequencyEveryOtherDay, true},
		{"weekly", FrequencyWeekly, true},
		{"invalid", "", false},
		{"", "", false},
		{"DAILY", "", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseFrequency(tt.input)
			if ok != tt.wantOk {
				t.Errorf("ParseFrequency() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.want {
				t.Errorf("ParseFrequency() = %v, want %v", got, tt.want)
			}
		})
	}
}
