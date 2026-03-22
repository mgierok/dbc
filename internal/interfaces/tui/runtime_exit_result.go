package tui

func (m *Model) exitResultOrDefault() RuntimeExitResult {
	if m == nil || m.exitResult.Action == 0 {
		return runtimeExitResultQuit()
	}
	return m.exitResult
}
