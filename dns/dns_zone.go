package dns

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
	Records map[RecordType][]Record
}

// Add adds a record to the zone
func (z *Zone) Add(r Record) {
	if z.Records == nil {
		z.Records = map[RecordType][]Record{}
	}

	z.Records[r.Type()] = append(z.Records[r.Type()], r)
}

// Remove removes a record from the zone
func (z *Zone) Remove(r Record) {
	if z.Records == nil {
		z.Records = map[RecordType][]Record{}
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

// // SubdomainRecords is a list of record for a subdomain
// type SubdomainRecords []Record

// // Add adds a record to the zone
// func (r *SubdomainRecords) Add(record Record) {
// 	if r == nil {
// 		r = &SubdomainRecords{}
// 	}

// 	// avoid to add 2 time the same record
// 	for i := range *r {
// 		if (*r)[i] == record {
// 			return
// 		}
// 	}

// 	*r = append(*r, record)
// }

// // Remove removes a record from the zone
// func (r *SubdomainRecords) Remove(record Record) {
// 	if r == nil {
// 		r = &SubdomainRecords{}
// 	}

// 	for i := range *r {
// 		if (*r)[i] == record {
// 			*r = append((*r)[:i], (*r)[i+1:]...)
// 		}
// 	}
// }

// func (r *SubdomainRecords) MarshalJSON() ([]byte, error) {
// 	output := make(map[string][]Record, len(*r))

// 	for i := range *r {
// 		record := (*r)[i]
// 		if _, ok := output[string(record.Type())]; !ok {
// 			output[string(record.Type())] = []Record{}
// 		}
// 		output[string(record.Type())] = append(output[string(record.Type())], record)
// 	}

// 	return json.Marshal(output)
// }

// func (r *SubdomainRecords) UnmarshalJSON(data []byte, v interface{}) error {

// 	sdr := v.(SubdomainRecords)

// 	input := make(map[string][]Record, len(*r))
// 	if err := json.Unmarshal(data, &input); err != nil {
// 		return err
// 	}
// 	for t, r := range input {
// 		switch t {
// 		case RecordA:

// 		}
// 	}

// 	return json.Marshal(output)
// }

type ZoneOwner struct {
	Owner string //threebot ID owning this zone
}
