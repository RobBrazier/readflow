package form

import (
	"errors"
	"maps"

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

func ValidationMinValues[T comparable](min int) func([]T) error {
	return func(t []T) error {
		if len(t) < min {
			return errors.New("You must select at least one")
		}
		return nil
	}
}

func ValidationRequired[T comparable]() func(T) error {
	return func(t T) error {
		var empty T
		if t == empty {
			return errors.New("This field is required")
		}
		return nil
	}
}

func Confirm(message string, value *bool) *huh.Confirm {
	return huh.NewConfirm().
		Title(message).
		Value(value)
}
