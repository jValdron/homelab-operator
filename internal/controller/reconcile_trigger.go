package controller

import "sync"

// reconcileTrigger identifies the watch event that enqueued a reconcile (best-effort when coalesced).
type reconcileTrigger struct {
	Kind string
	Name string
}

type triggerRecorder struct {
	mu    sync.Mutex
	byKey map[string]reconcileTrigger
}

func (t *triggerRecorder) note(key, kind, name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.byKey == nil {
		t.byKey = make(map[string]reconcileTrigger)
	}
	t.byKey[key] = reconcileTrigger{Kind: kind, Name: name}
}

func (t *triggerRecorder) consume(key string) (reconcileTrigger, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	tr, ok := t.byKey[key]
	if ok {
		delete(t.byKey, key)
	}
	return tr, ok
}

func triggerLogValues(tr reconcileTrigger) []any {
	if tr.Kind == "" {
		return nil
	}
	return []any{"triggerKind", tr.Kind, "trigger", tr.Name}
}
