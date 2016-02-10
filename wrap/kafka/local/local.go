package local

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/thatguystone/cog"
)

// A Handle ensures that Kafka is running. Be sure to call Handle.Close() when
// you no longer need Kafka.
type Handle struct {
	closed bool
}

const addr = "localhost:63445"

var (
	localDir = ""

	mtx     sync.Mutex
	refCnt  = 0
	dataDir = ""
	cmd     *exec.Cmd
)

func init() {
	_, abspath, _, _ := runtime.Caller(1)
	localDir = path.Dir(abspath)
}

// Run returns a handle to a running instances of Kafka. The instance is
// shared across calls, and terminated once all handles are closed.
func Run() *Handle {
	mtx.Lock()
	defer mtx.Unlock()

	if refCnt == 0 {
		out, err := exec.Command("gradle", "-p", localDir, "jar").CombinedOutput()
		cog.Must(err, "failed to build local kafka: %s", string(out))

		dataDir, err = ioutil.TempDir("", "localkafka")
		cog.Must(err, "failed to create tempdir")

		cmd = exec.Command("java",
			"-jar",
			filepath.Join(localDir, "build", "libs", "local.jar"),
			dataDir)
		err = cmd.Start()
		cog.Must(err, "failed to start local kafka")

		// In case it takes a while to start...
		ready := false
		for !ready {
			c, err := net.DialTimeout("tcp", addr, time.Millisecond*200)
			if err != nil {
				time.Sleep(time.Millisecond * 50)
			} else {
				ready = true
				c.Close()
			}
		}
	}

	refCnt++
	return &Handle{}
}

func decRef() {
	refCnt--
	if refCnt == 0 {
		err := cmd.Process.Kill()
		cog.Must(err, "failed to kill local kafka")

		cmd.Wait()
		cmd = nil

		err = os.RemoveAll(dataDir)
		cog.Must(err, "failed to clean up tmpdir: %s", dataDir)
		dataDir = ""
	}
}

// Close closes this handle so that Kafka can be shutdown when it's no longer
// needed.
func (h *Handle) Close() {
	mtx.Lock()
	defer mtx.Unlock()

	if !h.closed {
		decRef()
		h.closed = true
	}
}

// Addr gets the address to the running Kafka broker
func (h *Handle) Addr() string {
	return addr
}
