package domain

type Frequency string

const (
	FrequencyDaily        Frequency = "daily"
	FrequencyEveryOtherDay Frequency = "every_other_day"
	FrequencyWeekly       Frequency = "weekly"
)

func (f Frequency) String() string {
	return string(f)
}

func (f Frequency) DisplayName() string {
	switch f {
	case FrequencyDaily:
		return "Ежедневно"
	case FrequencyEveryOtherDay:
		return "Через день"
	case FrequencyWeekly:
		return "Раз в неделю"
	default:
		return string(f)
	}
}

func ParseFrequency(s string) (Frequency, bool) {
	switch s {
	case string(FrequencyDaily):
		return FrequencyDaily, true
	case string(FrequencyEveryOtherDay):
		return FrequencyEveryOtherDay, true
	case string(FrequencyWeekly):
		return FrequencyWeekly, true
	default:
		return "", false
	}
}
