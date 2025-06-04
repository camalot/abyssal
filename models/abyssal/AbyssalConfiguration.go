package abyssal


type AbyssalConfiguration struct {
	Version string `yaml:"version"`
	Packages []AbyssalPackage `yaml:"packages"`
}