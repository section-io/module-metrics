package metrics

import (
	"os"
	"testing"
)

func TestCreateLogFifo(t *testing.T) {
	path := "/tmp/test-file"
	expectedMode := "prw-rw-rw-"
	err := CreateLogFifo(path)
	if err != nil {
		t.Error(err)
	}

	fileinfo, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}

	mode := fileinfo.Mode()
	if mode.String() != expectedMode {
		t.Errorf("Mode of file %s is %s, not expected %s", path, mode, expectedMode)
	}

	os.Remove(path)
}

func TestCreateLogFifoFails(t *testing.T) {
	nonExistantPath := "/i/dont/exists/test-file"

	err := CreateLogFifo(nonExistantPath)

	if err == nil {
		t.Errorf("CreateLogFifo didn't fail when creating %s", nonExistantPath)
	}
}

//func TestInit(t *testing.T) {
//
//	Init(os.Stdout)
//
//	time.Sleep(time.Second * 1)
//
//	file, err := os.OpenFile(NamedPipePath, os.O_RDWR, os.ModeNamedPipe)
//	if err != nil {
//		log.Fatalf("error opening file: %v", err)
//	}
//
//	for i := 0; i < 5; i++ {
//		file.WriteString(`{"bytes":"2202","content_type":"text/html","hostname":"www.cjb.io","http_x_forwarded_proto":"http","referrer":"-","request":"GET / HTTP/1.1","response_body_bytes":"1811","section_io_id":"6d7da8ca4fae05d37f60849438c157b7","status":"200","time":"2018-09-27T19:17:20+0000","time_taken_ms":"0","upstream_addr":"-","upstream_status":"-","useragent":"curl/7.47.0"}` + "\n")
//	}
//
//	time.Sleep(time.Second * 1)
//
//}
