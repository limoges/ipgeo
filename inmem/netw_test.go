package inmem_test

import (
	"net"
	"testing"

	"github.com/limoges/ipgeo"
	"github.com/limoges/ipgeo/inmem"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNetworkLocator(t *testing.T) {
	fs := afero.NewMemMapFs()
	name := "csv"
	f, err := fs.Create(name)
	if err != nil {
		t.Error(err)
	}
	_, err = f.WriteString(`
1.0.0.0/24,2151718
1.0.1.0/24,1810821
1.0.2.0/23,1810821
1.0.4.0/22,2077456
1.0.8.0/21,1809858
`)
	if err != nil {
		t.Error(err)
	}

	testcases := []struct {
		addr net.IP
		id   ipgeo.LocationID
		err  bool
		msg  string
	}{
		{msg: "1", addr: net.IPv4(1, 0, 0, 0), id: 2151718, err: false},
		{msg: "2", addr: net.IPv4(1, 0, 1, 0), id: 1810821, err: false},
		{msg: "3", addr: net.IPv4(1, 0, 2, 0), id: 1810821, err: false},
		{msg: "4", addr: net.IPv4(1, 0, 4, 0), id: 2077456, err: false},
		{msg: "5", addr: net.IPv4(1, 0, 8, 0), id: 1809858, err: false},
	}

	locator, err := inmem.NewNetworkLocatorFromFs(fs, name)
	if err != nil {
		t.Error(err)
	}

	for _, testcase := range testcases {
		id, err := locator.FindNetworkLocation(testcase.addr)
		if testcase.err {
			assert.NotNil(t, err, testcase.msg)
		} else {
			assert.Nil(t, err, testcase.msg)
			assert.Equal(t, testcase.id, id, testcase.msg)
		}
	}
}
