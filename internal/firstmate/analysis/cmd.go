package analysis

import "os/exec"

// runCommand executes a command in the given directory and returns combined output.
func runCommand(cwd, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	return string(out), err
}
