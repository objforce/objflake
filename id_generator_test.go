package objflake

import "testing"

func TestNextId(t *testing.T) {
	objGenerator := NewIDGenerator()

	keyPrefix := []byte("abc")
	podIdentifier := []byte("def")

	orgId, err := objGenerator.NextID(keyPrefix, podIdentifier)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("orgId: %d, %v\n", len(orgId), orgId)
	orgId, err = objGenerator.NextID(keyPrefix, podIdentifier)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("orgId: %v\n", orgId)
}
