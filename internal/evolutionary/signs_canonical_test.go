package evolutionary

import (
	"reflect"
	"testing"

	"github.com/aj-nt/empirical/internal/zodiac"
)

func TestSignsCanonical(t *testing.T) {
	if !reflect.DeepEqual(signs, zodiac.Signs) {
		t.Fatal("evolutionary.signs != zodiac.Signs")
	}
}
