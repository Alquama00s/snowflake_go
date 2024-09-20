package snowflakego

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var snowflake *snowFlake

func init() {
	sfSet, err := newSnowflakeSettings(39, 8, 1e7)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	sf, err := newSnowflake(*sfSet, time.Date(2023, time.December, 0, 0, 0, 0, 0, time.UTC), 12)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	snowflake = sf
}

func TestIdGeneration(t *testing.T) {

	res := make(map[uint64]int)
	cnt := 100000
	for i := 0; i < cnt; i++ {
		id, err := snowflake.nextId()
		if err != nil {
			t.Fatal(err)
		}
		res[id] = i
		fmt.Println(id)
	}
	if len(res) != cnt {
		t.Fatal("expected ", cnt, " but found ", len(res))
		t.Failed()
	}
}
