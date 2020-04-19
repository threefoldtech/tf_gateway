package dns

// RecordType is the enum type for all supported DNS record
type RecordType string

// Enum value for RecordType
var (
	RecordTypeA     = RecordType("a")
	RecordTypeCNAME = RecordType("aaaa")
	RecordTypeAAA   = RecordType("cname")
)

// Record define the interface to be a DNS record
type Record interface {
	Type() RecordType
}

// RecordA is a type A DNS record
type RecordA struct {
	IP4 string `json:"ip4"`
	TTL int    `json:"ttl"`
}

// Type implements Record interface
func (r RecordA) Type() RecordType {
	return RecordTypeA
}

// RecordAAA is a type AAA DNS record
type RecordAAA struct {
	IP6 string `json:"ip6"`
	TTL int    `json:"ttl"`
}

// Type implements Record interface
func (r RecordAAA) Type() RecordType {
	return RecordTypeAAA
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
	Records map[string][]Record
	Owner   string //threebot ID owning this zone
}

// Add adds a record to the zone
func (z *Zone) Add(name string, r Record) {
	if z.Records == nil {
		z.Records = map[string][]Record{}
	}

	_, ok := z.Records[name]
	if !ok {
		z.Records[name] = []Record{r}
	} else {
		z.Records[name] = append(z.Records[name], r)
	}
}

// Remove removes a record from the zone
func (z *Zone) Remove(name string, r Record) {
	if z.Records == nil {
		z.Records = map[string][]Record{}
	}

	_, ok := z.Records[name]
	if !ok {
		return
	}

	for i := range z.Records[name] {
		if z.Records[name][i] == r {
			z.Records[name] = append(z.Records[name][:i], z.Records[name][i+1:]...)
		}
	}
}
