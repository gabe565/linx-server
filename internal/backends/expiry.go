package backends

import (
	"encoding/json"
	"time"
)

type Expiry time.Time

func (e *Expiry) UnmarshalJSON(b []byte) error {
	var t time.Time
	if err := t.UnmarshalJSON(b); err == nil {
		*e = Expiry(t)
		return nil
	}

	var unix int64
	if err := json.Unmarshal(b, &unix); err != nil {
		return err
	}

	if unix == 0 {
		*e = Expiry(time.Time{})
	} else {
		*e = Expiry(time.Unix(unix, 0))
	}
	return nil
}

func (e Expiry) MarshalJSON() ([]byte, error) {
	return time.Time(e).MarshalJSON()
}
