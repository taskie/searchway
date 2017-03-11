package srchway

type OperationType int

const (
	OperationTypeNone OperationType = iota
	OperationTypeSearch
	OperationTypeInfo
	OperationTypeGet
	OperationTypeHelp
	OperationTypeVersion
)

type Conf struct {
	Operation    OperationType
	Args         []string
	Verbose      bool
	AurFlag      bool
	OfficialFlag bool
	JsonFlag     bool
	MultilibFlag bool
	TestingFlag  bool
}

func (conf Conf) Repos() (repos []Repo) {
	repos = make([]Repo, 0)
	if conf.OfficialFlag {
		repos = append(repos, OfficialRepo{})
	}
	if conf.AurFlag {
		repos = append(repos, UserRepo{})
	}
	return
}
