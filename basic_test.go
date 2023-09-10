package bcl

import "testing"

const basicInput = `
var syncthing_port = 8384
var local_start    = 9384
var domain = "foo.org"
var z2 = z + 1
var z  = local_start + 1000
var s  = "sth" + 1 + "-" + domain
var cond  = true
var cond2 = false
var cond_big = not not not not cond


tunnel "hosty-syncthing" {
	local_port  = local_start
	remote_port = syncthing_port
	host = "hosty." + domain
	enabled = not cond2
}

tunnel "another-syncthing" {
	host = "yet" + "." + "another.com"
	local_port  = local_start + 1
	remote_port = syncthing_port
	some_field  = z2
	enabled  = cond
	prepared = true
	started  = cond_big
}
`

func basicRun() ([]Block, error) {
	return Interp([]byte(basicInput))
}

func TestBasic(t *testing.T) {
	_, err := basicRun()
	if err != nil {
		t.Error("unexpected error:", err)
	}
}

func BenchmarkBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		basicRun()
	}
}
