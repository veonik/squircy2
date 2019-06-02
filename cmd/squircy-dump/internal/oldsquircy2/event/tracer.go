package event

type Tracer struct {
	Limit int
	data  map[EventType][]map[string]interface{}
}

func NewTracer(evm EventManager) *Tracer {
	t := &Tracer{25, make(map[EventType][]map[string]interface{}, 0)}
	evm.Bind(AllEvents, func(ev Event) {
		history, ok := t.data[ev.Type]
		if !ok {
			history = make([]map[string]interface{}, 0)
		}

		if len(history) >= t.Limit {
			history = history[1:]
		}

		history = append(history, ev.Data)

		t.data[ev.Type] = history
	})

	return t
}

func (t *Tracer) History(evt EventType) []map[string]interface{} {
	if history, ok := t.data[evt]; ok {
		return history
	}

	return nil
}
