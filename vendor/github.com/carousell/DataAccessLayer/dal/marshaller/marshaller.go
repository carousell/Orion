//Package marshaller implements all marshallers used in DAL.
//
//marshaller provides a generic way to use same field for 'sql', 'gocql' and 'json' packages
package marshaller

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
)

/*
NullString represents a string that may be null. NullString implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.String
	} else {
	   // NULL value
	}
*/
type NullString struct {
	sql.NullString
}

//MarshalCQL implements the gocql.Marshaler interface.
func (v NullString) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, v.String)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (v *NullString) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	v.Valid = true
	return gocql.Unmarshal(info, data, &v.String)
}

//Scan implements the sql.Scanner interface.
func (v *NullString) Scan(src interface{}) error {
	return v.NullString.Scan(src)
}

//Value implements the driver.Valuer interface.
func (v NullString) Value() (driver.Value, error) {
	return v.NullString.Value()
}

//MarshalJSON implements json.Marshaler interface.
func (v NullString) MarshalJSON() (text []byte, err error) {
	if !v.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(v.String)
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (v *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &v.String)
	v.Valid = err == nil
	return err
}

/*
NullBool represents a boolean value that may be null. NullBool implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.Bool
	} else {
	   // NULL value
	}
*/
type NullBool struct {
	sql.NullBool
}

//MarshalCQL implements the gocql.Marshaler interface.
func (v NullBool) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, v.Bool)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (v *NullBool) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	v.Valid = true
	return gocql.Unmarshal(info, data, &v.Bool)
}

//Scan implements the sql.Scanner interface.
func (v *NullBool) Scan(src interface{}) error {
	return v.NullBool.Scan(src)
}

//Value implements the driver.Valuer interface.
func (v NullBool) Value() (driver.Value, error) {
	return v.NullBool.Value()
}

//MarshalJSON implements json.Marshaler interface.
func (v NullBool) MarshalJSON() (text []byte, err error) {
	return json.Marshal(v.Bool)
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (v *NullBool) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &v.Bool)
	v.Valid = err == nil
	return err
}

// This is an implementation where the json value of the bool
// is also null in case its not valid.
type NullJsonBool struct {
	NullBool
}

//MarshalJSON implements json.Marshaler interface.
func (v NullJsonBool) MarshalJSON() (text []byte, err error) {
	if !v.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(v.Bool)
}

/*
NullInt64 represents a integer value that may be null. NullInt64 implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.Int64
	} else {
	   // NULL value
	}
*/
type NullInt64 struct {
	sql.NullInt64
}

//MarshalCQL implements the gocql.Marshaler interface.
func (v NullInt64) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, v.Int64)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (v *NullInt64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	v.Valid = true
	return gocql.Unmarshal(info, data, &v.Int64)
}

//Scan implements the sql.Scanner interface.
func (v *NullInt64) Scan(src interface{}) error {
	return v.NullInt64.Scan(src)
}

//Value implements the driver.Valuer interface.
func (v NullInt64) Value() (driver.Value, error) {
	return v.NullInt64.Value()
}

//MarshalJSON implements json.Marshaler interface.
func (v NullInt64) MarshalJSON() (text []byte, err error) {
	return json.Marshal(v.Int64)
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (v *NullInt64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &v.Int64)
	v.Valid = err == nil
	return err
}

/*
NullFloat64 represents a float64/double value that may be null. NullFloat64 implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.Float64
	} else {
	   // NULL value
	}
*/
type NullFloat64 struct {
	sql.NullFloat64
}

//MarshalCQL implements the gocql.Marshaler interface.
func (v NullFloat64) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, v.Float64)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (v *NullFloat64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	v.Valid = true
	return gocql.Unmarshal(info, data, &v.Float64)
}

//Scan implements the sql.Scanner interface.
func (v *NullFloat64) Scan(src interface{}) error {
	return v.NullFloat64.Scan(src)
}

//Value implements the driver.Valuer interface.
func (v NullFloat64) Value() (driver.Value, error) {
	return v.NullFloat64.Value()
}

//MarshalJSON implements json.Marshaler interface.
func (v NullFloat64) MarshalJSON() (text []byte, err error) {
	return json.Marshal(v.Float64)
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (v *NullFloat64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &v.Float64)
	v.Valid = err == nil
	return err
}

/*
NullTime represents a time.Time value that may be null. NullTime implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.Time
	} else {
	   // NULL value
	}
NullTime trucates milliseconds to seconds
*/
type NullTime struct {
	Time  time.Time
	Valid bool
}

//Scan implements the sql.Scanner interface.
func (v *NullTime) Scan(src interface{}) error {
	v.Time, v.Valid = src.(time.Time)
	if v.Valid {
		v.Time = v.Time.In(time.UTC).Truncate(time.Millisecond)
	}
	return nil
}

//Value implements the driver.Valuer interface.
func (v NullTime) Value() (driver.Value, error) {
	if !v.Valid {
		return nil, nil
	}
	return v.Time, nil
}

//MarshalCQL implements the gocql.Marshaler interface.
func (v NullTime) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, v.Time)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (v *NullTime) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	err := gocql.Unmarshal(info, data, &v.Time)
	if err == nil {
		v.Time = v.Time.In(time.UTC).Truncate(time.Millisecond)
		v.Valid = true
	}
	return err
}

//MarshalJSON implements json.Marshaler interface.
func (v NullTime) MarshalJSON() (text []byte, err error) {
	return json.Marshal(v.Time)
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (v *NullTime) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &v.Time)
	if err == nil {
		v.Time = v.Time.In(time.UTC).Truncate(time.Millisecond)
		v.Valid = true
	}
	return err
}

/*
NullTimeWithMillis represents a time.Time value that may be null. NullTime implements multiple interfaces so it can be used in SQL/CQL/JSON
	if s.Valid {
	   // use s.Time
	} else {
	   // NULL value
	}
NullTimeWithMillis does not trucate milliseconds.
*/
type NullTimeWithMillis struct {
	Time  time.Time
	Valid bool
}

//MarshalCQL implements the gocql.Marshaler interface.
func (n NullTimeWithMillis) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return gocql.Marshal(info, n.Time)
}

//UnmarshalCQL implements gocql.Unmarshaler interface.
func (n *NullTimeWithMillis) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	err := gocql.Unmarshal(info, data, &n.Time)
	if err == nil {
		n.Time = n.Time.In(time.UTC).Truncate(time.Microsecond)
		n.Valid = true
	}
	return err
}

//UnmarshalJSON implements json.Unmarshaler interface.
func (n *NullTimeWithMillis) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &n.Time)
	if err == nil {
		n.Time = n.Time.In(time.UTC).Truncate(time.Microsecond)
		n.Valid = true
	}
	return err
}

//MarshalJSON implements json.Marshaler interface.
func (n NullTimeWithMillis) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Time)
}

//Scan implements the sql.Scanner interface.
func (n *NullTimeWithMillis) Scan(src interface{}) error {
	n.Time, n.Valid = src.(time.Time)
	if n.Valid {
		n.Time = n.Time.In(time.UTC).Truncate(time.Microsecond)
	}
	return nil
}

//Value implements the driver.Valuer interface.
func (n NullTimeWithMillis) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}
