package hashmap

import "hash/fnv"

type HashMap struct {
	slots      []entry
	size       int
	tombstones int
}

type entry struct {
	key   string
	value int
	state uint8 // 0: empty, 1: occupied, 2: tombstone
}

func New() *HashMap {
	return &HashMap{
		slots: make([]entry, 8),
	}
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return h.Sum64()
}

func (h *HashMap) hash(key string) int {
	return int(hashKey(key) % uint64(len(h.slots)))
}

func (h *HashMap) Put(key string, value int) {
	index := h.hash(key)
	firstTombstone := -1

	for {
		e := &h.slots[index]

		if e.state == 0 {
			target := index

			if firstTombstone != -1 {
				target = firstTombstone
				h.tombstones--
			}

			h.slots[target] = entry{key: key, value: value, state: 1}
			h.size++
			break
		}

		if e.state == 1 && e.key == key {
			e.value = value
			return
		}

		if e.state == 2 && firstTombstone == -1 {
			firstTombstone = index
		}

		index = (index + 1) % len(h.slots)
	}

	if (h.size+h.tombstones)*4 >= len(h.slots)*3 {
		h.resize()
	}
}

func (h *HashMap) Get(key string) (int, bool) {
	if len(h.slots) == 0 {
		return 0, false
	}

	index := h.hash(key)

	for i := 0; i < len(h.slots); i++ {
		e := &h.slots[index]

		if e.state == 0 {
			return 0, false
		}

		if e.state == 1 && e.key == key {
			return e.value, true
		}

		index = (index + 1) % len(h.slots)
	}

	return 0, false
}

func (h *HashMap) Delete(key string) bool {
	if len(h.slots) == 0 {
		return false
	}

	index := h.hash(key)

	for i := 0; i < len(h.slots); i++ {
		e := &h.slots[index]

		if e.state == 0 {
			return false
		}

		if e.state == 1 && e.key == key {
			e.state = 2
			h.size--
			h.tombstones++
			return true
		}

		index = (index + 1) % len(h.slots)
	}

	return false
}

func (h *HashMap) Len() int {
	return h.size
}

func (h *HashMap) resize() {
	old := h.slots
	newSlots := make([]entry, len(old)*2)

	for _, e := range old {
		if e.state != 1 {
			continue
		}

		index := int(hashKey(e.key) % uint64(len(newSlots)))

		for newSlots[index].state == 1 {
			index = (index + 1) % len(newSlots)
		}

		newSlots[index] = e
	}

	h.slots = newSlots
	h.tombstones = 0
}
