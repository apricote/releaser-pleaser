package updater

type ReleaseInfo struct {
	Version        string
	ChangelogEntry string
}

type Updater func(string) (string, error)

type NewUpdater func(ReleaseInfo) Updater

func WithInfo(info ReleaseInfo, constructors ...NewUpdater) []Updater {
	updaters := make([]Updater, 0, len(constructors))
	for _, constructor := range constructors {
		updaters = append(updaters, constructor(info))
	}

	return updaters
}
