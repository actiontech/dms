package channel

func IsClosed(ch <-chan struct{}) bool {
	if ch == nil {
		return true
	}
	select {
	case _, ok := <-ch:
		if !ok {
			return true
		}
	default:
	}

	return false
}

func TryClose(ch chan struct{}) {
	if !IsClosed(ch) {
		close(ch)
	}
}
