package cloudian

import (
	"fmt"
	"testing"
)

func TestRenderTerraBytesAsKiloBytes(t *testing.T) {
	expected := fmt.Sprintf("%d", 3*1024*1024*1024)

	if actual := (3 * TB).KBString(); actual != expected {
		t.Errorf("Expected 3 TB expressed in KB to be %s, got %s", expected, actual)
	}
}
