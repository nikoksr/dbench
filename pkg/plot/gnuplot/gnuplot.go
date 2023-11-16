package gnuplot

import (
	"context"
	"os"
	"os/exec"
)

func ExecuteScript(ctx context.Context, script string) error {
	cmd := exec.CommandContext(ctx, "gnuplot")
	cmd.Stderr = os.Stderr

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, err = stdinPipe.Write([]byte(script))
	if err != nil {
		return err
	}
	stdinPipe.Close()

	return cmd.Wait()
}
