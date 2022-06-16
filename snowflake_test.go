package snowflake

import (
	"sync"
	"testing"
)

var wg sync.WaitGroup
var muList sync.Mutex = sync.Mutex{}

func TestId(t *testing.T) {
	var n = 0
	var list []int64 = []int64{}
	for {
		n++
		if n == 100000 {
			break
		}
		// wg.Add(1)
		// go func() {
		// defer wg.Done()
		ids := ID.NextId()
		t.Log(ids)
		// muList.Lock()
		list = append(list, ids)
		// muList.Unlock()
		// }()

		//time.Sleep(1000 * 1e6)

	}
	// wg.Wait()
	tempMap := map[int64]byte{}
	for _, e := range list {
		tempMap[e] = 0
	}
	t.Log(len(tempMap))
	t.Log(len(list))

}
