package bicache_test

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jamiealquiza/bicache/v2"
)

// Benchmarks

func BenchmarkGet(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
		c.Set(k, "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Get(keys[i])
	}
}

func BenchmarkSet(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Set(keys[i], "my value")
	}
}

func BenchmarkSetTTL(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.SetTTL(keys[i], "my value", 3600)
	}
}

func BenchmarkDel(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	for i := 0; i < b.N; i++ {
		c.Set(keys[i], "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Del(keys[i])
	}
}

func BenchmarkList(b *testing.B) {
	b.StopTimer()

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10000,
		MRUSize:    600000,
		ShardCount: 1024,
		AutoEvict:  30000,
	})

	keys := make([]string, b.N/2)
	for i := 0; i < b.N/2; i++ {
		k := strconv.Itoa(i)
		keys[i] = k
	}

	for i := 0; i < b.N/2; i++ {
		c.Set(keys[i], "my value")
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = c.List(b.N / 2)
	}
}

// Tests

func TestSetGet(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  10000,
	})

	ok := c.Set("key", "value")
	if !ok {
		t.Error("Set failed")
	}

	if c.Get("key") != "value" {
		t.Error("Get failed")
	}

	// Ensure that updates work.
	ok = c.Set("key", "value2")
	if !ok {
		t.Error("Update failed")
	}

	if c.Get("key") != "value2" {
		t.Error("Update failed")
	}
}

func TestSetTTL(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	ok := c.SetTTL("key", "value", 3)
	if !ok {
		t.Error("Set failed")
	}

	if c.Get("key") != "value" {
		t.Error("Get failed")
	}

	log.Printf("Sleeping for 4 seconds to allow evictions")
	time.Sleep(4 * time.Second)

	if c.Get("key") != nil {
		t.Error("Key TTL expiration failed")
	}
}

func TestDel(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  10000,
	})

	c.Set("key", "value")
	c.Del("key")

	if c.Get("key") != nil {
		t.Error("Delete failed")
	}

	stats := c.Stats()

	if stats.MRUSize != 0 {
		t.Errorf("Expected MRU size 0, got %d", stats.MRUSize)
	}
}

func TestList(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	c.Get("0")
	c.Get("0")
	c.Get("0")
	c.Get("0")

	c.Get("1")
	c.Get("1")
	c.Get("1")

	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")

	log.Printf("Sleeping for 2 seconds to allow evictions")
	time.Sleep(2 * time.Second)

	list := c.List(5)

	if len(list) != 5 {
		t.Errorf("Exected list output len of 5, got %d", len(list))
	}

	// Can only reliably expect the MFU nodes
	// in the list output. Check that the top3
	// are what's expected.
	expected := []string{"2", "0", "1"}
	for i, n := range list[:3] {
		if n.Key != expected[i] {
			t.Errorf(`Expected key "%s" at list element %d, got "%s"`,
				expected[i], i, n.Key)
		}
	}
}

func TestFlushMRU(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 30; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	// Check before.
	stats := c.Stats()
	if stats.MRUSize != 30 {
		t.Errorf("Expected MFU size of 30, got %d", stats.MFUSize)
	}

	c.FlushMRU()

	// Check after.
	stats = c.Stats()
	if stats.MRUSize != 0 {
		t.Errorf("Expected MRU size of 0, got %d", stats.MRUSize)
	}
}

func TestFlushMFU(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	c.Get("0")
	c.Get("0")
	c.Get("0")
	c.Get("0")

	c.Get("1")
	c.Get("1")
	c.Get("1")

	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")

	log.Printf("Sleeping for 2 seconds to allow promotions")
	time.Sleep(2 * time.Second)

	// MFU promotion is already tested in bicache_tests.go
	// TestPromoteEvict. This is somewhat of a dupe, but
	// ensure that if we're testing for a length of 0
	// after a flush that it wasn't 0 to begin with.

	// Check before.
	stats := c.Stats()
	if stats.MFUSize != 3 {
		t.Errorf("Expected MFU size of 3, got %d", stats.MFUSize)
	}

	c.FlushMFU()

	// Check after.
	stats = c.Stats()
	if stats.MFUSize != 0 {
		t.Errorf("Expected MFU size of 0, got %d", stats.MFUSize)
	}
}

func TestFlushAll(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 1,
		AutoEvict:  1000,
	})

	for i := 0; i < 40; i++ {
		c.Set(strconv.Itoa(i), "value")
	}

	c.Get("0")
	c.Get("0")
	c.Get("0")
	c.Get("0")

	c.Get("1")
	c.Get("1")
	c.Get("1")

	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")
	c.Get("2")

	log.Printf("Sleeping for 2 seconds to allow promotions")
	time.Sleep(2 * time.Second)

	// Check before.
	stats := c.Stats()

	if stats.MFUSize != 3 {
		t.Errorf("Expected MFU size of 3, got %d", stats.MFUSize)
	}

	if stats.MRUSize != 30 {
		t.Errorf("Expected MFU size of 30, got %d", stats.MFUSize)
	}

	c.FlushAll()

	// Check after.
	stats = c.Stats()

	if stats.MFUSize != 0 {
		t.Errorf("Expected MFU size of 0, got %d", stats.MFUSize)
	}

	if stats.MRUSize != 0 {
		t.Errorf("Expected MRU size of 0, got %d", stats.MRUSize)
	}
}

func TestIntegrity(t *testing.T) {
	words := []string{"&c", "'d", "'em", "'ll", "'m", "'mid", "'midst", "'mongst", "'prentice", "'re", "'s", "'sblood", "'sbodikins", "'sdeath", "'sfoot", "'sheart", "'shun", "'slid", "'slife", "'slight", "'snails", "'strewth", "'t", "'til", "'tis", "'twas", "'tween", "'twere", "'twill", "'twixt", "'twould", "'un", "'ve", "1080", "10th", "1st", "2", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "a", "a'", "a's", "a/c", "a1", "aa", "aaa", "aah", "aahed", "aahing", "aahs", "aal", "aalii", "aaliis", "aals", "aam", "aardvark", "aardvarks", "aardwolf", "aardwolves", "aargh", "aaron", "aaronic", "aarrgh", "aarrghh", "aas", "aasvogel", "aasvogels", "ab", "aba", "abac", "abaca", "abacas", "abacate", "abacaxi", "abacay", "abaci", "abacinate", "abacination", "abacisci", "abaciscus", "abacist", "aback", "abacli", "abacot", "abacterial", "abactinal", "abactinally", "abaction", "abactor", "abaculi", "abaculus", "abacus", "abacuses", "abada", "abaddon", "abadejo", "abadengo", "abadia", "abaff", "abaft", "abaisance", "abaised", "abaiser", "abaisse", "abaissed", "abaka", "abakas", "abalation", "abalienate", "abalienated", "abalienating", "abalienation", "abalone", "abalones", "abamp", "abampere", "abamperes", "abamps", "aband", "abandon", "abandonable", "abandoned", "abandonedly", "abandonee", "abandoner", "abandoners", "abandoning", "abandonment", "abandonments", "abandons", "abandum", "abanet", "abanga", "abannition", "abapical", "abaptiston", "abaptistum", "abarthrosis", "abarticular", "abarticulation", "abas", "abase", "abased", "abasedly", "abasedness", "abasement", "abasements", "abaser", "abasers", "abases", "abash", "abashed", "abashedly", "abashedness", "abashes", "abashing", "abashless", "abashlessly", "abashment", "abashments", "abasia", "abasias", "abasic", "abasing", "abasio", "abask", "abassi", "abastard", "abastardize", "abastral", "abatable", "abatage", "abate", "abated", "abatement", "abatements", "abater", "abaters", "abates", "abatic", "abating", "abatis", "abatised", "abatises", "abatjour", "abatjours", "abaton", "abator", "abators", "abattage", "abattis", "abattised", "abattises", "abattoir", "abattoirs", "abattu", "abattue", "abature", "abaue", "abave", "abaxial", "abaxile", "abay", "abayah", "abaze", "abb", "abba", "abbacies", "abbacomes", "abbacy", "abbandono", "abbas", "abbasi", "abbasid", "abbassi", "abbate", "abbatial", "abbatical", "abbatie", "abbaye", "abbe", "abbes", "abbess", "abbesses", "abbest", "abbevillian", "abbey", "abbey's", "abbeys", "abbeystead", "abbeystede", "abboccato", "abbogada", "abbot", "abbot's", "abbotcies", "abbotcy", "abbotnullius", "abbotric", "abbots", "abbotship", "abbotships", "abbott", "abbozzo", "abbr", "abbrev", "abbreviatable", "abbreviate", "abbreviated", "abbreviately", "abbreviates", "abbreviating", "abbreviation", "abbreviations", "abbreviator", "abbreviators", "abbreviatory", "abbreviature", "abbroachment", "abby", "abc", "abcess", "abcissa", "abcoulomb", "abd", "abdal", "abdali", "abdaria", "abdat", "abdest", "abdicable", "abdicant", "abdicate", "abdicated", "abdicates", "abdicating", "abdication", "abdications", "abdicative", "abdicator", "abditive", "abditory", "abdom", "abdomen", "abdomen's", "abdomens", "abdomina", "abdominal", "abdominales", "abdominalia", "abdominalian", "abdominally", "abdominals", "abdominoanterior", "abdominocardiac", "abdominocentesis", "abdominocystic", "abdominogenital", "abdominohysterectomy", "abdominohysterotomy", "abdominoposterior", "abdominoscope", "abdominoscopy", "abdominothoracic", "abdominous", "abdominovaginal", "abdominovesical", "abduce", "abduced", "abducens", "abducent", "abducentes", "abduces", "abducing", "abduct", "abducted", "abducting", "abduction", "abduction's", "abductions", "abductor", "abductor's", "abductores", "abductors", "abducts", "abeam", "abear", "abearance", "abecedaire", "abecedaria", "abecedarian", "abecedarians", "abecedaries", "abecedarium", "abecedarius", "abecedary", "abed", "abede", "abedge", "abegge", "abeigh", "abel", "abele", "abeles", "abelian", "abelite", "abelmosk", "abelmosks", "abelmusk", "abeltree", "abend", "abends", "abenteric", "abepithymia", "aberdavine", "aberdeen", "aberdevine", "aberduvine", "abernethy", "aberr", "aberrance", "aberrancies", "aberrancy", "aberrant", "aberrantly", "aberrants", "aberrate", "aberrated", "aberrating", "aberration", "aberrational", "aberrations", "aberrative", "aberrator", "aberrometer", "aberroscope", "aberuncate", "aberuncator", "abesse", "abessive", "abet", "abetment", "abetments", "abets", "abettal", "abettals", "abetted", "abetter", "abetters", "abetting", "abettor", "abettors", "abevacuation", "abey", "abeyance", "abeyances", "abeyancies", "abeyancy", "abeyant", "abfarad", "abfarads", "abhenries", "abhenry", "abhenrys", "abhinaya", "abhiseka", "abhominable", "abhor", "abhorred", "abhorrence", "abhorrences", "abhorrency", "abhorrent", "abhorrently", "abhorrer", "abhorrers", "abhorrible", "abhorring", "abhors", "abib", "abichite", "abidal", "abidance", "abidances", "abidden", "abide", "abided", "abider", "abiders", "abides", "abidi", "abiding", "abidingly", "abidingness", "abiegh", "abience", "abient", "abietate", "abietene", "abietic", "abietin", "abietineous", "abietinic", "abietite", "abigail", "abigails", "abigailship", "abigeat", "abigei", "abigeus", "abilao", "abilene", "abiliment", "abilitable", "abilities", "ability", "ability's", "abilla", "abilo", "abime", "abintestate", "abiogeneses", "abiogenesis", "abiogenesist", "abiogenetic", "abiogenetical", "abiogenetically", "abiogenist", "abiogenous", "abiogeny", "abiological", "abiologically", "abiology", "abioses", "abiosis", "abiotic", "abiotical", "abiotically", "abiotrophic", "abiotrophy", "abir", "abirritant", "abirritate", "abirritated", "abirritating", "abirritation", "abirritative", "abiston", "abit", "abiuret", "abject", "abjectedness", "abjection", "abjections"}

	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 8,
		AutoEvict:  2000,
	})

	for _, w := range words {
		c.Set(w, w)
	}

	// Test pre-eviction integrity.
	for _, w := range words {
		if v := c.Get(w); v != nil && v != w {
			t.Errorf("Expected value %s, got %s", v, w)
		}
	}

	time.Sleep(3 * time.Second)

	// Test post-eviction integrity.
	for _, w := range words {
		if v := c.Get(w); v != nil && v != w {
			t.Errorf("Expected value %s, got %s", v, w)
		}
	}

	c.FlushAll()

	// Test integrity for keys that
	// traverse MFU/MRU.

	// Add item targeted for MFU.
	c.Set("promoted", "promoted")
	for i := 0; i < 20; i++ {
		c.Get("promoted")
	}

	// Seq scan to exhaust MRU.
	for _, w := range words {
		c.Set(w, w)
	}

	time.Sleep(3 * time.Second)

	// Check that target item
	// was promoted.

	if c.Get("promoted") != "promoted" {
		t.Errorf(`Expected MFU item value "promoted", got "%s"`, c.Get("promoted"))
	}

	// Check MFU key state.
	keys := c.List(len(words))
	for _, k := range keys {
		if k.Key == "promoted" && k.State != 1 {
			t.Errorf("Expected key state 1, got %d", k.State)
		}
	}

	c, _ = bicache.New(&bicache.Config{
		MFUSize:    1,
		MRUSize:    30,
		ShardCount: 1,
		AutoEvict:  2000,
	})

	// Add item targeted for MFU
	// promotion then demotion.
	c.Set("promoted", "promoted")
	for i := 0; i < 5; i++ {
		c.Get("promoted")
	}

	// Seq scan to exhaust MRU.
	for _, w := range words {
		c.Set(w, w)
	}

	time.Sleep(3 * time.Second)

	// Check that the promoted key
	// has the proper key state.
	keys = c.List(len(words))
	for _, k := range keys {
		if k.State == 1 && k.Key != "promoted" {
			t.Errorf(`Expected key name "promoted", got "%s"`, k.Key)
		}
	}

	// Set MFU replacement target.
	c.Set("replacement", "replacement")
	for i := 0; i < 10; i++ {
		c.Get("replacement")
	}

	// Seq scan to exhaust MRU.
	for _, w := range words {
		c.Set(w, w)
	}

	time.Sleep(3 * time.Second)

	// Check that the promoted key
	// has the proper key state.
	keys = c.List(len(words))
	for _, k := range keys {
		if k.State == 1 && k.Key != "replacement" {
			t.Errorf(`Expected key name "replacement", got "%s"`, k.Key)
		}
	}

	// Check that the demoted key
	// has the proper key state.
	keys = c.List(len(words))
	for _, k := range keys {
		if k.Key == "promoted" && k.State != 0 {
			t.Errorf(`Expected key "promoted" to be in state 0, got "%d"`, k.State)
		}
	}

	// Previous and new MFU keys
	// should still be present.
	if c.Get("replacement") != "replacement" || c.Get("promoted") != "promoted" {
		t.Errorf("Unexpected cache miss")
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	c, _ := bicache.New(&bicache.Config{
		MFUSize:    10,
		MRUSize:    30,
		ShardCount: 2,
		AutoEvict:  1000,
	})

	const numTasks = 10000
	wg := &sync.WaitGroup{}
	wg.Add(numTasks * 2)

	for i := 0; i < numTasks; i++ {
		go func() {
			defer wg.Done()
			v := fmt.Sprint(rand.Int())
			ok := c.SetTTL("key", v, 3)
			if !ok {
				t.Error("Set failed")
			}
		}()

		go func() {
			defer wg.Done()
			c.Get("key")
		}()
	}
	wg.Wait()
}
