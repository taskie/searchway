package srchway

type Repo interface {
	Search(conf Conf) (bytes []byte, err error)
	Info(conf Conf) (bytes []byte, err error)
	Get(conf Conf) (newOutFilePath string, err error)
	PrintSearchResponse(conf Conf) (err error)
	PrintInfoResponse(conf Conf) (err error)
}
