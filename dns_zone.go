package tfgateway

// RecordType is the enum type for all supported DNS record
type RecordType string

// Enum value for RecordType
var (
	RecordTypeA     = RecordType("a")
	RecordTypeCNAME = RecordType("aaaa")
	RecordTypeAAA   = RecordType("cname")
)

// DNSRecord
type DNSRecord interface {
	Type() RecordType
}

type RecordA struct {
	IP4 string `json:"ip4"`
	TTL int    `json:"ttl"`
}

func (r RecordA) Type() RecordType {
	return RecordTypeA
}

type RecordAAA struct {
	IP6 string `json:"ip6"`
	TTL int    `json:"ttl"`
}

func (r RecordAAA) Type() RecordType {
	return RecordTypeAAA
}

type RecordCname struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
}

func (r RecordCname) Type() RecordType {
	return RecordTypeCNAME
}

type Zone struct {
	Records map[string][]DNSRecord
	Owner   string //threebot ID owning this zone
}

func (z *Zone) AddRecord(name string, r DNSRecord) {
	if z.Records == nil {
		z.Records = map[string][]DNSRecord{}
	}

	_, ok := z.Records[name]
	if !ok {
		z.Records[name] = []DNSRecord{r}
	} else {
		z.Records[name] = append(z.Records[name], r)
	}
}

func (z *Zone) RemoveRecord(name string, r DNSRecord) {
	if z.Records == nil {
		z.Records = map[string][]DNSRecord{}
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
