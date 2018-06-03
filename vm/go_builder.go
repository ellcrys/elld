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

// GetRunScript run the blockcode execution script
func (gb *goBuilder) GetRunScript() []string {
	cmd := []string{"bash", "-c", "/bin/bcode"}
	return cmd
}

// GetBuildScript returns build script
func (gb *goBuilder) GetBuildScript() []string {
	path := makeCopyPath(gb.containerID)
	script := `cd ` + path + ` && go build -v -o /bin/bcode && ls /bin/bcode`
	buildCmd := []string{"bash", "-c", "echo >> __script.sh '" + script + "' && bash ./__script.sh"}
	return buildCmd
}
