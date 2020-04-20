package internal

func antiCheatRunner() StageRunner {
	return StageRunner{
		isDebug: false,
		stages:  []Stage{},
	}
}
