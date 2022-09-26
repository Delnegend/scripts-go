package libs

import(
	"os"
	"os/exec"
)

func ExecCmd(logfile *os.File, program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}