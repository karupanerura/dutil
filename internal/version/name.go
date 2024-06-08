package version

import "strconv"

var Name string

func init() {
	var name []byte
	name = strconv.AppendInt(name, Major, 10)
	name = append(name, '.')
	name = strconv.AppendInt(name, Minor, 10)
	name = append(name, '.')
	name = strconv.AppendInt(name, Patch, 10)
	Name = string(name)
}
