package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/threefoldtech/zos/pkg/provision"
	"github.com/threefoldtech/zos/pkg/versioned"
)

var (
	// reservationSchemaV1 reservation schema version 1
	reservationSchemaV1 = versioned.MustParse("1.0.0")
	// ReservationSchemaLastVersion link to latest version
	reservationSchemaLastVersion = reservationSchemaV1
)

const _reservationsKey = "tfgateway_reservations"

// Redis is a in reservation cache using the filesystem as backend
type Redis struct {
	sync.RWMutex
	pool *redis.Pool
}

// NewRedis creates a in memory reservation store
func NewRedis(pool *redis.Pool) *Redis {
	return &Redis{
		pool: pool,
	}
}

// Sync update the statser with all the reservation present in the cache
func (s *Redis) Sync(statser provision.Statser) error {
	//this should probably be reversed and moved to the Statser object instead
	s.RLock()
	defer s.RUnlock()

	con := s.pool.Get()
	defer con.Close()

	ids, err := redis.ByteSlices(con.Do("HKEYS", _reservationsKey))
	if err != nil {
		return err
	}

	for _, id := range ids {
		r, err := s.get(string(id))
		if err != nil {
			return err
		}
		statser.Increment(r)
	}

	return nil
}

// Add a reservation to the store
func (s *Redis) Add(r *provision.Reservation) error {
	s.Lock()
	defer s.Unlock()

	con := s.pool.Get()
	defer con.Close()

	buf := bytes.Buffer{}
	writer, err := versioned.NewWriter(&buf, reservationSchemaLastVersion)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(writer).Encode(r); err != nil {
		return err
	}

	_, err = con.Do("HSET", _reservationsKey, r.ID, buf.Bytes())
	return err
}

// Remove a reservation from the store
func (s *Redis) Remove(id string) error {
	s.Lock()
	defer s.Unlock()

	con := s.pool.Get()
	defer con.Close()

	_, err := con.Do("HDEL", _reservationsKey, id)
	return err
}

// GetExpired returns all id the the reservations that are expired
// at the time of the function call
func (s *Redis) GetExpired() ([]*provision.Reservation, error) {
	s.RLock()
	defer s.RUnlock()

	con := s.pool.Get()
	defer con.Close()

	ids, err := redis.ByteSlices(con.Do("HKEYS", _reservationsKey))
	if err != nil {
		return nil, err
	}

	rs := make([]*provision.Reservation, 0, len(ids))
	for _, id := range ids {
		r, err := s.get(string(id))
		if err != nil {
			return nil, err
		}

		if r.Expired() {
			// r.Tag = Tag{"source": "FSStore"}
			rs = append(rs, r)
		}
	}

	return rs, nil
}

// Get retrieves a specific reservation using its ID
// if returns a non nil error if the reservation is not present in the store
func (s *Redis) Get(id string) (*provision.Reservation, error) {
	s.RLock()
	defer s.RUnlock()

	return s.get(id)
}

// getType retrieves a specific reservation's type using its ID
// if returns a non nil error if the reservation is not present in the store
func (s *Redis) getType(id string) (provision.ReservationType, error) {
	r, err := s.get(id)
	if err != nil {
		return provision.ReservationType(0), err
	}
	return r.Type, nil
}

// Exists checks if the reservation ID is in the store
func (s *Redis) Exists(id string) (bool, error) {
	s.RLock()
	defer s.RUnlock()

	con := s.pool.Get()
	defer con.Close()

	return redis.Bool(con.Do("HEXISTS", _reservationsKey, id))
}

func (s *Redis) get(id string) (*provision.Reservation, error) {
	con := s.pool.Get()
	defer con.Close()

	b, err := redis.Bytes(con.Do("HGET", _reservationsKey, id))
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)
	reader, err := versioned.NewReader(r)
	if versioned.IsNotVersioned(err) {
		r = bytes.NewReader(b)
		reader = versioned.NewVersionedReader(versioned.MustParse("0.0.0"), r)
	}

	validV1 := versioned.MustParseRange(fmt.Sprintf("<=%s", reservationSchemaV1))
	var reservation provision.Reservation

	if validV1(reader.Version()) {
		if err := json.NewDecoder(reader).Decode(&reservation); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unknown reservation object version (%s)", reader.Version())
	}
	// reservation.Tag = Tag{"source": "FSStore"}
	return &reservation, nil
}

// Close makes sure the backend of the store is closed properly
func (s *Redis) Close() error {
	return nil
}
