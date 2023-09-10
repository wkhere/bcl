package bcl

import "testing"

const basicInput = `
var syncthing_port = 8384
var local_start    = 9384
var domain = "foo.org"
var z0 = -z1 + +0      # forward decl
var z1 = local_start - syncthing_port
var z2 = -1 - (---10/2-1) + z1
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
	enabled  = cond
	prepared = true
	started  = cond_big
	u = 4+3*2
	v = z0
	x = z2 - z1
	y = 3 - ---8/4 - (+1-3/2)
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
