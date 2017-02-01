package srchway

type PrintMode int

const (
	NormalMode PrintMode = iota
	JsonMode
)

type Repo interface {
	Search(query string) (bytes []byte, err error)
	Info(query string) (bytes []byte, err error)
	Get(query string) (err error)
	PrintSearchResponse(query string, mode PrintMode) (err error)
	PrintInfoResponse(query string, mode PrintMode) (err error)
}
