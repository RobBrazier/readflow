package form

import (
	"errors"
	"maps"
	"strconv"

	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/charmbracelet/huh"
)

var sourceLabels = map[string]string{
	"database": "Calibre + Calibre-Web databases",
}

var targetLabels = map[string]string{
	"anilist":   "Anilist.co",
	"hardcover": "Hardcover.app",
}

func getOptionLabel(key string, labels map[string]string) string {
	label, ok := labels[key]
	if ok {
		return label
	}
	return key
}

func SourceSelect(value *string) *huh.Select[string] {
	sources := source.GetSources()
	options := []huh.Option[string]{}
	for name := range maps.Keys(sources) {
		label := getOptionLabel(name, sourceLabels)
		option := huh.NewOption(label, name)
		options = append(options, option)
	}
	return huh.NewSelect[string]().
		Options(options...).
		Title("Enabled Source").
		Description("Where should we get the reading data from?").
		Value(value)
}

func TargetSelect(value *[]string) *huh.MultiSelect[string] {
	targets := *target.GetTargets()
	options := []huh.Option[string]{}
	for _, target := range targets {
		name := target.GetName()
		label := getOptionLabel(name, targetLabels)
		option := huh.NewOption(label, name)

		options = append(options, option)
	}
	return huh.NewMultiSelect[string]().
		Options(options...).
		Title("Enabled Sync Targets").
		Description("Where do you your reading updates to be sent to?").
		Validate(ValidationMinValues[string](1)).
		Value(value)
}

type intAccessor struct {
	Value *int
}

func (ia intAccessor) Get() string {
	return strconv.Itoa(*ia.Value)
}

func (ia intAccessor) Set(value string) {
	val, err := strconv.Atoi(value)
	if err == nil {
		return
	}
	*ia.Value = val
}

func SyncDays(value *int) *huh.Input {
	if *value == 0 {
		*value = 1
	}

	strValue := strconv.Itoa(*value)

	return huh.NewInput().
		Title("Sync Days").
		Description("How many days do you want to look at when syncing?").
		Validate(func(s string) error {
			if err := ValidationRequired[string]()(s); err != nil {
				return err
			}
			if val, err := strconv.Atoi(s); err != nil || val < 1 || val > 30 {
				return errors.New("Please specify a number between 1 and 30")
			}
			return nil
		}).
		Accessor(intAccessor{Value: value}).
		Value(&strValue)
}

func Confirm(message string, value *bool) *huh.Confirm {
	return huh.NewConfirm().
		Title(message).
		Value(value)
}
