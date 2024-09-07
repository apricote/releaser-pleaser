package releasepr

// Label is the string identifier of a pull/merge request label on the forge.
type Label struct {
	Color       string
	Name        string
	Description string
}

var (
	LabelNextVersionTypeNormal = Label{
		Color:       "EFC15B",
		Name:        "rp-next-version::normal",
		Description: "Request a stable version",
	}
	LabelNextVersionTypeRC = Label{
		Color:       "EFC15B",
		Name:        "rp-next-version::rc",
		Description: "Request a pre-release -rc version",
	}
	LabelNextVersionTypeBeta = Label{
		Color:       "EFC15B",
		Name:        "rp-next-version::beta",
		Description: "Request a pre-release -beta version",
	}
	LabelNextVersionTypeAlpha = Label{
		Color:       "EFC15B",
		Name:        "rp-next-version::alpha",
		Description: "Request a pre-release -alpha version",
	}
)

var (
	LabelReleasePending = Label{
		Color:       "DEDEDE",
		Name:        "rp-release::pending",
		Description: "Release for this PR is pending",
	}
	LabelReleaseTagged = Label{
		Color:       "0E8A16",
		Name:        "rp-release::tagged",
		Description: "Release for this PR is created",
	}
)

var KnownLabels = []Label{
	LabelNextVersionTypeNormal,
	LabelNextVersionTypeRC,
	LabelNextVersionTypeBeta,
	LabelNextVersionTypeAlpha,

	LabelReleasePending,
	LabelReleaseTagged,
}
