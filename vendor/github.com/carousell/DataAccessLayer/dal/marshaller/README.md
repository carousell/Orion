# marshaller
`import "./marshaller"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package marshaller implements all marshallers used in DAL.

marshaller provides a generic way to use same field for 'sql', 'gocql' and 'json' packages

## <a name="pkg-imports">Imported Packages</a>

- github.com/gocql/gocql

## <a name="pkg-index">Index</a>
* [type NullBool](#NullBool)
  * [func (v NullBool) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullBool.MarshalCQL)
  * [func (v NullBool) MarshalJSON() (text []byte, err error)](#NullBool.MarshalJSON)
  * [func (v \*NullBool) Scan(src interface{}) error](#NullBool.Scan)
  * [func (v \*NullBool) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullBool.UnmarshalCQL)
  * [func (v \*NullBool) UnmarshalJSON(b []byte) error](#NullBool.UnmarshalJSON)
  * [func (v NullBool) Value() (driver.Value, error)](#NullBool.Value)
* [type NullFloat64](#NullFloat64)
  * [func (v NullFloat64) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullFloat64.MarshalCQL)
  * [func (v NullFloat64) MarshalJSON() (text []byte, err error)](#NullFloat64.MarshalJSON)
  * [func (v \*NullFloat64) Scan(src interface{}) error](#NullFloat64.Scan)
  * [func (v \*NullFloat64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullFloat64.UnmarshalCQL)
  * [func (v \*NullFloat64) UnmarshalJSON(b []byte) error](#NullFloat64.UnmarshalJSON)
  * [func (v NullFloat64) Value() (driver.Value, error)](#NullFloat64.Value)
* [type NullInt64](#NullInt64)
  * [func (v NullInt64) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullInt64.MarshalCQL)
  * [func (v NullInt64) MarshalJSON() (text []byte, err error)](#NullInt64.MarshalJSON)
  * [func (v \*NullInt64) Scan(src interface{}) error](#NullInt64.Scan)
  * [func (v \*NullInt64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullInt64.UnmarshalCQL)
  * [func (v \*NullInt64) UnmarshalJSON(b []byte) error](#NullInt64.UnmarshalJSON)
  * [func (v NullInt64) Value() (driver.Value, error)](#NullInt64.Value)
* [type NullJsonBool](#NullJsonBool)
  * [func (v NullJsonBool) MarshalJSON() (text []byte, err error)](#NullJsonBool.MarshalJSON)
* [type NullString](#NullString)
  * [func (v NullString) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullString.MarshalCQL)
  * [func (v NullString) MarshalJSON() (text []byte, err error)](#NullString.MarshalJSON)
  * [func (v \*NullString) Scan(src interface{}) error](#NullString.Scan)
  * [func (v \*NullString) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullString.UnmarshalCQL)
  * [func (v \*NullString) UnmarshalJSON(b []byte) error](#NullString.UnmarshalJSON)
  * [func (v NullString) Value() (driver.Value, error)](#NullString.Value)
* [type NullTime](#NullTime)
  * [func (v NullTime) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullTime.MarshalCQL)
  * [func (v NullTime) MarshalJSON() (text []byte, err error)](#NullTime.MarshalJSON)
  * [func (v \*NullTime) Scan(src interface{}) error](#NullTime.Scan)
  * [func (v \*NullTime) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullTime.UnmarshalCQL)
  * [func (v \*NullTime) UnmarshalJSON(b []byte) error](#NullTime.UnmarshalJSON)
  * [func (v NullTime) Value() (driver.Value, error)](#NullTime.Value)
* [type NullTimeWithMillis](#NullTimeWithMillis)
  * [func (n NullTimeWithMillis) MarshalCQL(info gocql.TypeInfo) ([]byte, error)](#NullTimeWithMillis.MarshalCQL)
  * [func (n NullTimeWithMillis) MarshalJSON() ([]byte, error)](#NullTimeWithMillis.MarshalJSON)
  * [func (n \*NullTimeWithMillis) Scan(src interface{}) error](#NullTimeWithMillis.Scan)
  * [func (n \*NullTimeWithMillis) UnmarshalCQL(info gocql.TypeInfo, data []byte) error](#NullTimeWithMillis.UnmarshalCQL)
  * [func (n \*NullTimeWithMillis) UnmarshalJSON(b []byte) error](#NullTimeWithMillis.UnmarshalJSON)
  * [func (n NullTimeWithMillis) Value() (driver.Value, error)](#NullTimeWithMillis.Value)

#### <a name="pkg-files">Package files</a>
[marshaller.go](./marshaller.go) 

## <a name="NullBool">type</a> [NullBool](./marshaller.go#L71-L73)
``` go
type NullBool struct {
    sql.NullBool
}
```
NullBool represents a boolean value that may be null. NullBool implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.Bool
	} else {
	   // NULL value
	}

### <a name="NullBool.MarshalCQL">func</a> (NullBool) [MarshalCQL](./marshaller.go#L76)
``` go
func (v NullBool) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullBool.MarshalJSON">func</a> (NullBool) [MarshalJSON](./marshaller.go#L97)
``` go
func (v NullBool) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullBool.Scan">func</a> (\*NullBool) [Scan](./marshaller.go#L87)
``` go
func (v *NullBool) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullBool.UnmarshalCQL">func</a> (\*NullBool) [UnmarshalCQL](./marshaller.go#L81)
``` go
func (v *NullBool) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullBool.UnmarshalJSON">func</a> (\*NullBool) [UnmarshalJSON](./marshaller.go#L102)
``` go
func (v *NullBool) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullBool.Value">func</a> (NullBool) [Value](./marshaller.go#L92)
``` go
func (v NullBool) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

## <a name="NullFloat64">type</a> [NullFloat64](./marshaller.go#L175-L177)
``` go
type NullFloat64 struct {
    sql.NullFloat64
}
```
NullFloat64 represents a float64/double value that may be null. NullFloat64 implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.Float64
	} else {
	   // NULL value
	}

### <a name="NullFloat64.MarshalCQL">func</a> (NullFloat64) [MarshalCQL](./marshaller.go#L180)
``` go
func (v NullFloat64) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullFloat64.MarshalJSON">func</a> (NullFloat64) [MarshalJSON](./marshaller.go#L201)
``` go
func (v NullFloat64) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullFloat64.Scan">func</a> (\*NullFloat64) [Scan](./marshaller.go#L191)
``` go
func (v *NullFloat64) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullFloat64.UnmarshalCQL">func</a> (\*NullFloat64) [UnmarshalCQL](./marshaller.go#L185)
``` go
func (v *NullFloat64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullFloat64.UnmarshalJSON">func</a> (\*NullFloat64) [UnmarshalJSON](./marshaller.go#L206)
``` go
func (v *NullFloat64) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullFloat64.Value">func</a> (NullFloat64) [Value](./marshaller.go#L196)
``` go
func (v NullFloat64) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

## <a name="NullInt64">type</a> [NullInt64](./marshaller.go#L130-L132)
``` go
type NullInt64 struct {
    sql.NullInt64
}
```
NullInt64 represents a integer value that may be null. NullInt64 implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.Int64
	} else {
	   // NULL value
	}

### <a name="NullInt64.MarshalCQL">func</a> (NullInt64) [MarshalCQL](./marshaller.go#L135)
``` go
func (v NullInt64) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullInt64.MarshalJSON">func</a> (NullInt64) [MarshalJSON](./marshaller.go#L156)
``` go
func (v NullInt64) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullInt64.Scan">func</a> (\*NullInt64) [Scan](./marshaller.go#L146)
``` go
func (v *NullInt64) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullInt64.UnmarshalCQL">func</a> (\*NullInt64) [UnmarshalCQL](./marshaller.go#L140)
``` go
func (v *NullInt64) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullInt64.UnmarshalJSON">func</a> (\*NullInt64) [UnmarshalJSON](./marshaller.go#L161)
``` go
func (v *NullInt64) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullInt64.Value">func</a> (NullInt64) [Value](./marshaller.go#L151)
``` go
func (v NullInt64) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

## <a name="NullJsonBool">type</a> [NullJsonBool](./marshaller.go#L110-L112)
``` go
type NullJsonBool struct {
    NullBool
}
```
This is an implementation where the json value of the bool
is also null in case its not valid.

### <a name="NullJsonBool.MarshalJSON">func</a> (NullJsonBool) [MarshalJSON](./marshaller.go#L115)
``` go
func (v NullJsonBool) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

## <a name="NullString">type</a> [NullString](./marshaller.go#L23-L25)
``` go
type NullString struct {
    sql.NullString
}
```
NullString represents a string that may be null. NullString implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.String
	} else {
	   // NULL value
	}

### <a name="NullString.MarshalCQL">func</a> (NullString) [MarshalCQL](./marshaller.go#L28)
``` go
func (v NullString) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullString.MarshalJSON">func</a> (NullString) [MarshalJSON](./marshaller.go#L49)
``` go
func (v NullString) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullString.Scan">func</a> (\*NullString) [Scan](./marshaller.go#L39)
``` go
func (v *NullString) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullString.UnmarshalCQL">func</a> (\*NullString) [UnmarshalCQL](./marshaller.go#L33)
``` go
func (v *NullString) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullString.UnmarshalJSON">func</a> (\*NullString) [UnmarshalJSON](./marshaller.go#L57)
``` go
func (v *NullString) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullString.Value">func</a> (NullString) [Value](./marshaller.go#L44)
``` go
func (v NullString) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

## <a name="NullTime">type</a> [NullTime](./marshaller.go#L221-L224)
``` go
type NullTime struct {
    Time  time.Time
    Valid bool
}
```
NullTime represents a time.Time value that may be null. NullTime implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.Time
	} else {
	   // NULL value
	}

NullTime trucates milliseconds to seconds

### <a name="NullTime.MarshalCQL">func</a> (NullTime) [MarshalCQL](./marshaller.go#L244)
``` go
func (v NullTime) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullTime.MarshalJSON">func</a> (NullTime) [MarshalJSON](./marshaller.go#L259)
``` go
func (v NullTime) MarshalJSON() (text []byte, err error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullTime.Scan">func</a> (\*NullTime) [Scan](./marshaller.go#L227)
``` go
func (v *NullTime) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullTime.UnmarshalCQL">func</a> (\*NullTime) [UnmarshalCQL](./marshaller.go#L249)
``` go
func (v *NullTime) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullTime.UnmarshalJSON">func</a> (\*NullTime) [UnmarshalJSON](./marshaller.go#L264)
``` go
func (v *NullTime) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullTime.Value">func</a> (NullTime) [Value](./marshaller.go#L236)
``` go
func (v NullTime) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

## <a name="NullTimeWithMillis">type</a> [NullTimeWithMillis](./marshaller.go#L282-L285)
``` go
type NullTimeWithMillis struct {
    Time  time.Time
    Valid bool
}
```
NullTimeWithMillis represents a time.Time value that may be null. NullTime implements multiple interfaces so it can be used in SQL/CQL/JSON

	if s.Valid {
	   // use s.Time
	} else {
	   // NULL value
	}

NullTimeWithMillis does not trucate milliseconds.

### <a name="NullTimeWithMillis.MarshalCQL">func</a> (NullTimeWithMillis) [MarshalCQL](./marshaller.go#L288)
``` go
func (n NullTimeWithMillis) MarshalCQL(info gocql.TypeInfo) ([]byte, error)
```
MarshalCQL implements the gocql.Marshaler interface.

### <a name="NullTimeWithMillis.MarshalJSON">func</a> (NullTimeWithMillis) [MarshalJSON](./marshaller.go#L313)
``` go
func (n NullTimeWithMillis) MarshalJSON() ([]byte, error)
```
MarshalJSON implements json.Marshaler interface.

### <a name="NullTimeWithMillis.Scan">func</a> (\*NullTimeWithMillis) [Scan](./marshaller.go#L318)
``` go
func (n *NullTimeWithMillis) Scan(src interface{}) error
```
Scan implements the sql.Scanner interface.

### <a name="NullTimeWithMillis.UnmarshalCQL">func</a> (\*NullTimeWithMillis) [UnmarshalCQL](./marshaller.go#L293)
``` go
func (n *NullTimeWithMillis) UnmarshalCQL(info gocql.TypeInfo, data []byte) error
```
UnmarshalCQL implements gocql.Unmarshaler interface.

### <a name="NullTimeWithMillis.UnmarshalJSON">func</a> (\*NullTimeWithMillis) [UnmarshalJSON](./marshaller.go#L303)
``` go
func (n *NullTimeWithMillis) UnmarshalJSON(b []byte) error
```
UnmarshalJSON implements json.Unmarshaler interface.

### <a name="NullTimeWithMillis.Value">func</a> (NullTimeWithMillis) [Value](./marshaller.go#L327)
``` go
func (n NullTimeWithMillis) Value() (driver.Value, error)
```
Value implements the driver.Valuer interface.

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)