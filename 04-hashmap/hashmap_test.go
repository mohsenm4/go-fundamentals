package hashmap

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	h := New()

	if h == nil {
		t.Fatal("New() returned nil")
	}

	if len(h.slots) != 8 {
		t.Fatalf("expected 8 slots, got %d", len(h.slots))
	}

	if h.size != 0 {
		t.Fatalf("expected size 0, got %d", h.size)
	}
}

func TestPutAndGet(t *testing.T) {
	h := New()

	h.Put("ali", 10)
	h.Put("reza", 20)
	h.Put("mohsen", 30)

	tests := []struct {
		key      string
		expected int
	}{
		{"ali", 10},
		{"reza", 20},
		{"mohsen", 30},
	}

	for _, tt := range tests {
		got, ok := h.Get(tt.key)

		if !ok {
			t.Fatalf("Get(%q) returned false", tt.key)
		}

		if got != tt.expected {
			t.Fatalf(
				"Get(%q) = %d, want %d",
				tt.key,
				got,
				tt.expected,
			)
		}
	}
}

func TestGetMissingKey(t *testing.T) {
	h := New()

	h.Put("ali", 10)

	got, ok := h.Get("reza")

	if ok {
		t.Fatal("Get() returned true for missing key")
	}

	if got != 0 {
		t.Fatalf("Get() returned %d, want 0", got)
	}
}

func TestPutUpdatesExistingKey(t *testing.T) {
	h := New()

	h.Put("ali", 10)
	h.Put("ali", 100)

	got, ok := h.Get("ali")

	if !ok {
		t.Fatal("key should exist")
	}

	if got != 100 {
		t.Fatalf("Get(ali) = %d, want 100", got)
	}

	if h.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", h.Len())
	}
}

func TestDelete(t *testing.T) {
	h := New()

	h.Put("ali", 10)
	h.Put("reza", 20)

	deleted := h.Delete("ali")

	if !deleted {
		t.Fatal("Delete(ali) returned false")
	}

	_, ok := h.Get("ali")

	if ok {
		t.Fatal("deleted key still exists")
	}

	got, ok := h.Get("reza")

	if !ok {
		t.Fatal("reza should still exist")
	}

	if got != 20 {
		t.Fatalf("Get(reza) = %d, want 20", got)
	}

	if h.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", h.Len())
	}
}

func TestDeleteMissingKey(t *testing.T) {
	h := New()

	h.Put("ali", 10)

	deleted := h.Delete("reza")

	if deleted {
		t.Fatal("Delete(reza) returned true for missing key")
	}

	if h.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", h.Len())
	}
}

func TestDeleteTwice(t *testing.T) {
	h := New()

	h.Put("ali", 10)

	if !h.Delete("ali") {
		t.Fatal("first Delete() returned false")
	}

	if h.Delete("ali") {
		t.Fatal("second Delete() returned true")
	}

	if h.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", h.Len())
	}
}

func TestCollision(t *testing.T) {
	h := New()

	// پیدا کردن دو key که به یک bucket می‌روند.
	key1 := "key1"
	key2 := ""

	index := h.hash(key1)

	for i := 0; ; i++ {
		candidate := "collision"

		if i > 0 {
			candidate += string(rune('a' + i%26))
		}

		if candidate != key1 && h.hash(candidate) == index {
			key2 = candidate
			break
		}
	}

	h.Put(key1, 100)
	h.Put(key2, 200)

	got1, ok1 := h.Get(key1)
	got2, ok2 := h.Get(key2)

	if !ok1 {
		t.Fatalf("Get(%q) failed", key1)
	}

	if !ok2 {
		t.Fatalf("Get(%q) failed", key2)
	}

	if got1 != 100 {
		t.Fatalf("Get(%q) = %d, want 100", key1, got1)
	}

	if got2 != 200 {
		t.Fatalf("Get(%q) = %d, want 200", key2, got2)
	}
}

func TestResize(t *testing.T) {
	h := New()

	for i := 0; i < 100; i++ {
		h.Put(fmt.Sprintf("key%d", i), i)
	}

	if len(h.slots) <= 8 {
		t.Fatalf("map did not resize, slots = %d", len(h.slots))
	}

	if h.Len() != 100 {
		t.Fatalf("Len() = %d, want 100", h.Len())
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)

		got, ok := h.Get(key)

		if !ok {
			t.Fatalf("Get(%q) failed after resize", key)
		}

		if got != i {
			t.Fatalf(
				"Get(%q) = %d, want %d",
				key,
				got,
				i,
			)
		}
	}
}

func TestDeleteAfterResize(t *testing.T) {
	h := New()

	for i := 0; i < 100; i++ {
		h.Put(fmt.Sprintf("key%d", i), i)
	}

	if !h.Delete("key50") {
		t.Fatal("Delete() failed")
	}

	_, ok := h.Get("key50")

	if ok {
		t.Fatal("deleted key still exists")
	}

	if h.Len() != 99 {
		t.Fatalf("Len() = %d, want 99", h.Len())
	}

	// مطمئن شویم بقیه keyها هنوز هستند.
	for i := 0; i < 100; i++ {
		if i == 50 {
			continue
		}

		key := fmt.Sprintf("key%d", i)

		got, ok := h.Get(key)

		if !ok {
			t.Fatalf("Get(%q) failed", key)
		}

		if got != i {
			t.Fatalf("Get(%q) = %d, want %d", key, got, i)
		}
	}
}

// probe chain نباید با حذف یک عنصر وسط بشکنه.
// سه کلید که به یک bucket می‌روند، وسطی را حذف می‌کنیم،
// سومی باید همچنان قابل پیدا کردن باشد (tombstone probe را قطع نمی‌کند).
func TestProbeChainAfterDelete(t *testing.T) {
	h := New()

	index := h.hash("anchor")

	keys := []string{"anchor"}

	for i := 0; len(keys) < 3; i++ {
		candidate := fmt.Sprintf("probe%d", i)

		if h.hash(candidate) == index {
			keys = append(keys, candidate)
		}
	}

	for i, k := range keys {
		h.Put(k, (i+1)*10)
	}

	if !h.Delete(keys[1]) {
		t.Fatalf("Delete(%q) failed", keys[1])
	}

	got, ok := h.Get(keys[2])

	if !ok {
		t.Fatalf("Get(%q) failed — probe chain broke after delete", keys[2])
	}

	if got != 30 {
		t.Fatalf("Get(%q) = %d, want 30", keys[2], got)
	}
}

// بعد از Delete و Put دوباره‌ی همان کلید، tombstone باید دوباره استفاده شود.
func TestReuseTombstone(t *testing.T) {
	h := New()

	h.Put("ali", 10)

	if !h.Delete("ali") {
		t.Fatal("Delete(ali) failed")
	}

	if h.tombstones != 1 {
		t.Fatalf("tombstones = %d, want 1", h.tombstones)
	}

	h.Put("ali", 20)

	if h.tombstones != 0 {
		t.Fatalf("tombstones = %d after reuse, want 0", h.tombstones)
	}

	if h.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", h.Len())
	}

	got, ok := h.Get("ali")

	if !ok || got != 20 {
		t.Fatalf("Get(ali) = (%d, %v), want (20, true)", got, ok)
	}
}

// چرخه‌ی طولانی Put/Delete نباید table را با tombstone پر کند
// (resize باید tombstoneها را پاک کند و probe بی‌نهایت نشود).
func TestNoTombstoneAccumulation(t *testing.T) {
	h := New()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("k%d", i)
		h.Put(key, i)
		h.Delete(key)
	}

	if h.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", h.Len())
	}

	h.Put("final", 42)

	got, ok := h.Get("final")

	if !ok || got != 42 {
		t.Fatalf("Get(final) = (%d, %v), want (42, true)", got, ok)
	}
}

func TestZeroValue(t *testing.T) {
	var h HashMap

	if h.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", h.Len())
	}
}
