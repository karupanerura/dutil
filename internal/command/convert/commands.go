package convert

type Commands struct {
	Table TableCommand `cmd:""`
	Key   KeyCommand   `cmd:""`
}
