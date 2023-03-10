package resourcepolicies

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const currentSupportDataVersion = "v1"

type csiVolumeSource struct {
	Driver string `yaml:"driver,omitempty"`
}

type nFSVolumeSource struct {
	// Server is the hostname or IP address of the NFS server
	Server string `yaml:"server,omitempty"`
	// Path is the exported NFS share
	Path string `yaml:"path,omitempty"`
}

// volumeConditions defined the current format of conditions we parsed
type volumeConditions struct {
	Capacity     string           `yaml:"capacity,omitempty"`
	StorageClass []string         `yaml:"storageClass,omitempty"`
	NFS          *nFSVolumeSource `yaml:"nfs,omitempty"`
	CSI          *csiVolumeSource `yaml:"csi,omitempty"`
}

func (p *policyConditionsMatcher) Validate() (bool, error) {
	for _, policy := range p.policies {
		for _, con := range policy.volConditions {
			valid, err := con.Validate()
			if err != nil {
				return false, errors.Wrap(err, "failed to validate conditions config")
			}
			if !valid {
				return false, nil
			}
		}
	}
	return true, nil
}

func (c *capacityCondition) Validate() (bool, error) {
	// [0, a]
	// [a, b]
	// [b, 0]
	// ==> low <= upper or upper is zero
	if (c.capacity.upper.Cmp(c.capacity.lower) >= 0) ||
		(!c.capacity.lower.IsZero() && c.capacity.upper.IsZero()) {
		return true, nil
	}
	return false, errors.Errorf("illegal values for capacity %v", c.capacity)

}

func (s *storageClassCondition) Validate() (bool, error) {
	// validate by yamlv3
	return true, nil
}

func (c *nfsCondition) Validate() (bool, error) {
	// validate by yamlv3
	return true, nil
}

func (c *csiCondition) Validate() (bool, error) {
	// validate by yamlv3
	return true, nil
}

// decodeStruct restric validate the keys in decoded mappings to exist as fields in the struct being decoded into
func decodeStruct(r io.Reader, s interface{}) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	return dec.Decode(s)
}

// Validate check action format
func (a *Action) Validate() (bool, error) {
	// Validate Type
	if a.Type != Skip {
		return false, fmt.Errorf("invalid action type %s", a.Type)
	}

	// TODO validate parameters
	return true, nil
}

// Validate func validate the key and value format of the user generated resource policies configuration
func Validate(resPolicies *ResourcePolicies) (bool, error) {
	if resPolicies.Version != currentSupportDataVersion {
		return false, fmt.Errorf("incompatible version number %s with supported version %s", resPolicies.Version, currentSupportDataVersion)
	}

	matcher := policyConditionsMatcher{}
	for k := range resPolicies.VolumePolicies {
		if ok, err := resPolicies.VolumePolicies[k].Action.Validate(); !ok {
			return false, errors.Wrap(err, "failed to validate config")
		}
		if err := matcher.addPolicy(&resPolicies.VolumePolicies[k]); err != nil {
			return false, errors.Wrap(err, "failed to validate config")
		}
	}

	return matcher.Validate()
}
