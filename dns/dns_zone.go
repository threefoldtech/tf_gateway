package dns

import "encoding/json"

// RecordType is the enum type for all supported DNS record
type RecordType string

// Enum value for RecordType
var (
	RecordTypeA     = RecordType("a")
	RecordTypeAAAA  = RecordType("aaaa")
	RecordTypeCNAME = RecordType("cname")
)

// Record define the interface to be a DNS record
type Record interface {
	Type() RecordType
}

// RecordA is a type A DNS record
type RecordA struct {
	IP4 string `json:"ip"`
	TTL int    `json:"ttl"`
}

// Type implements Record interface
func (r RecordA) Type() RecordType {
	return RecordTypeA
}

// RecordAAAA is a type AAAA DNS record
type RecordAAAA struct {
	IP6 string `json:"ip"`
	TTL int    `json:"ttl"`
}

// Type implements Record interface
func (r RecordAAAA) Type() RecordType {
	return RecordTypeAAAA
}

// RecordCname is a type CNAME DNS record
type RecordCname struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
}

// Type implements Record interface
func (r RecordCname) Type() RecordType {
	return RecordTypeCNAME
}

// Zone is a DNS zone. It hosts multiple records and belong to a owner
type Zone struct {
	Records records
}

// Add adds a record to the zone
func (z *Zone) Add(r Record) {
	if z.Records == nil {
		z.Records = records{}
	}

	z.Records[r.Type()] = append(z.Records[r.Type()], r)
}

// Remove removes a record from the zone
func (z *Zone) Remove(r Record) {
	if z.Records == nil {
		z.Records = records{}
	}

	_, ok := z.Records[r.Type()]
	if !ok {
		return
	}

	for i := range z.Records[r.Type()] {
		if z.Records[r.Type()][i] == r {
			z.Records[r.Type()] = append(z.Records[r.Type()][:i], z.Records[r.Type()][i+1:]...)
		}
	}
}

type records map[RecordType][]Record

// UnmarshalJSON implements encoding/json.Unmarshaler interface
func (rs records) UnmarshalJSON(b []byte) error {
	if rs == nil {
		rs = map[RecordType][]Record{}
	}

	m := make(map[RecordType][]json.RawMessage)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	for typ, records := range m {
		for _, b := range records {
			var r Record
			switch typ {
			case RecordTypeA:
				x := RecordA{}
				if err := json.Unmarshal(b, &x); err != nil {
					return err
				}
				r = x
			case RecordTypeAAAA:
				x := RecordAAAA{}
				if err := json.Unmarshal(b, &x); err != nil {
					return err
				}
				r = x
			case RecordTypeCNAME:
				x := RecordCname{}
				if err := json.Unmarshal(b, &x); err != nil {
					return err
				}
				r = x
			}
			rs[typ] = append(rs[typ], r)
		}
	}
	return nil
}

type ZoneOwner struct {
	Owner string //threebot ID owning this zone
}
