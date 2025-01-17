package git

type Repo struct {
	dir string

	MaxDays int
}

func NewRepo(dir string) *Repo {
	return &Repo{
		dir: dir,

		MaxDays: 60,
	}
}
