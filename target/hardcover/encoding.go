package hardcover

import (
	"encoding/json"
	"errors"
	"time"
)

func UnmarshalHardcoverDate(b []byte, v *time.Time) error {
	var input string
	json.Unmarshal(b, &input)
	parsedTime, err := time.Parse(time.DateOnly, input)
	if err != nil {
		return err
	}
	*v = parsedTime
	return nil
}

func MarshalHardcoverDate(v *time.Time) ([]byte, error) {
	if v == nil {
		return nil, errors.New("nil time value")
	}

	formattedTime := v.Format(time.DateOnly)
	return json.Marshal(formattedTime)
}
