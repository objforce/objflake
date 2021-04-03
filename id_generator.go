package objflake

import (
	"errors"
)

var ErrInvalidLengthKeyPrefix = errors.New("keyPrefix should be 3 characters")
var ErrInvalidLengthPodIdentifier = errors.New("keyPrefix should be 3 characters")

type IDGenerator struct {
	gen *Objflake
}

func NewIDGenerator() *IDGenerator {
	var st Settings
	return &IDGenerator{
		gen: NewObjflake(st),
	}
}

/**
 * 自增
 */
func (g *IDGenerator) NextID(keyPrefix []byte, podIdentifier []byte) (string, error) {
	if len(keyPrefix) != 3 {
		return "", ErrInvalidLengthKeyPrefix
	}

	if len(keyPrefix) != 3 {
		return "", ErrInvalidLengthPodIdentifier
	}

	numericIdentifier, err := g.gen.NextID()
	if err != nil {
		return "", err
	}

	newID := make([]byte, 15)
	copy(newID[0:3], keyPrefix)
	copy(newID[3:5], podIdentifier)

	sid, err := Encode(numericIdentifier)
	if err != nil {
		return "", err
	}
	copy(newID[5:15], sid)

	newID = computeEighteen(newID)

	return string(newID), nil
}
