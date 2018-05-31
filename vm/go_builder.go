package vm

type goBuilder struct {
	containerID string
}

// create new instance of goBuilder
func newGoBuilder(containerID string) *goBuilder {
	return &goBuilder{
		containerID: containerID,
	}
}

// get the run script that executes a blockcode
func (gb *goBuilder) GetRunScript() []string {
	cmd := []string{"bash", "-c", "/bin/bcode"}
	return cmd
}

// build a block code
func (gb *goBuilder) Build() []string {
	path := makeCopyPath(gb.containerID)
	execCmd := "cd " + path + " && go build -x -o /bin/bcode"
	buildCmd := []string{"bash", "-c", execCmd}
	return buildCmd
}
