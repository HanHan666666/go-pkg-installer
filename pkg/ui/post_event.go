package ui

var uiEventQueue = make(chan func(), 256)

// postUIEvent enqueues a UI update to run on the main UI loop.
func postUIEvent(fn func(), canDrop bool) {
	if fn == nil {
		return
	}
	if canDrop {
		select {
		case uiEventQueue <- fn:
		default:
		}
		return
	}
	uiEventQueue <- fn
}

func drainUIEvents(max int) {
	if max <= 0 {
		max = len(uiEventQueue)
	}
	for i := 0; i < max; i++ {
		select {
		case fn := <-uiEventQueue:
			if fn != nil {
				fn()
			}
		default:
			return
		}
	}
}
