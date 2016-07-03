package speed

import (
	"errors"
	"sync"
)

// InstanceDomain defines the interface for an instance domain
type InstanceDomain interface {
	ID() uint32                        // unique identifier for the instance domain
	Name() string                      // name of the instance domain
	Description() string               // description for the instance domain
	AddInstance(name string) error     // adds an instance to the indom
	AddInstances(names []string) error // adds multiple instances to the indom
	HasInstance(name string) bool      // checks if an instance is in the indom
	InstanceCount() int                // returns the number of instances in the indom
}

// PCPInstanceDomainBitLength is the maximum bit length of a PCP Instance Domain
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPInstanceDomainBitLength = 22

// PCPInstanceDomain wraps a PCP compatible instance domain
type PCPInstanceDomain struct {
	sync.RWMutex
	id                                uint32
	name                              string
	instances                         map[string]*pcpInstance
	offset                            int
	instanceOffset                    int
	shortDescription, longDescription *PCPString
}

// NewPCPInstanceDomain creates a new instance domain or returns an already created one for the passed name
// NOTE: this is different from parfait's idea of generating ids for InstanceDomains
// We simply generate a unique 32 bit hash for an instance domain name, and if it has not
// already been created, we create it, otherwise we return the already created version
func NewPCPInstanceDomain(name, shortDescription, longDescription string) (*PCPInstanceDomain, error) {
	if name == "" {
		return nil, errors.New("Instance Domain name cannot be empty")
	}

	return &PCPInstanceDomain{
		id:               getHash(name, PCPInstanceDomainBitLength),
		name:             name,
		instances:        make(map[string]*pcpInstance),
		shortDescription: NewPCPString(shortDescription),
		longDescription:  NewPCPString(longDescription),
	}, nil
}

// HasInstance returns true if an instance of the specified name is in the Indom
func (indom *PCPInstanceDomain) HasInstance(name string) bool {
	indom.RLock()
	defer indom.RUnlock()

	_, present := indom.instances[name]
	return present
}

// AddInstance adds a new instance to the current PCPInstanceDomain
func (indom *PCPInstanceDomain) AddInstance(name string) error {
	indom.Lock()
	defer indom.Unlock()

	if indom.HasInstance(name) {
		return errors.New("Instance with same name already created for the InstanceDomain")
	}

	ins := newpcpInstance(getHash(name, 0), name, indom)
	indom.instances[name] = ins

	return nil
}

// AddInstances adds new instances to the current PCPInstanceDomain
func (indom *PCPInstanceDomain) AddInstances(names []string) error {
	for _, name := range names {
		err := indom.AddInstance(name)
		if err != nil {
			return err
		}
	}
	return nil
}

// ID returns the id for PCPInstanceDomain
func (indom *PCPInstanceDomain) ID() uint32 { return indom.id }

// Name returns the name for PCPInstanceDomain
func (indom *PCPInstanceDomain) Name() string { return indom.name }

// InstanceCount returns the number of instances in the current instance domain
func (indom *PCPInstanceDomain) InstanceCount() int {
	indom.RLock()
	defer indom.RUnlock()

	return len(indom.instances)
}

// Description returns the description for PCPInstanceDomain
func (indom *PCPInstanceDomain) Description() string {
	s, l := indom.shortDescription.val, indom.longDescription.val
	if l != "" {
		return s + "\n\n" + l
	}
	return s
}

func (indom *PCPInstanceDomain) String() string {
	indom.RLock()
	defer indom.RUnlock()

	s := "InstanceDomain: " + indom.name
	if len(indom.instances) > 0 {
		s += "["
		for _, i := range indom.instances {
			s += i.name + ","
		}
		s += "]"
	}
	return s
}
