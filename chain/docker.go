package chain

import (
	"bytes"
	"os"
	"text/template"

	_ "embed"
)

//go:embed templates/docker-compose.yml.tpl
var dockerComposeTpl []byte

type DockerComposeConfig struct {
	Image   string
	Debug   bool
	LogFile string
	ChainID string
	Nodes   []NodeConfig

	Network DockerComposeNetwork
}

type DockerComposeNetwork struct {
	Subnet string
}

func GenerateDockerCompose(
	chainID string,
	image string,
	subnet string,
	nodes []NodeConfig,
	outDir string,
	debug bool,
) {
	cfg := &DockerComposeConfig{
		Image:   image,
		Debug:   debug,
		LogFile: "injectived.log",
		ChainID: chainID,
		Nodes:   nodes,
		Network: DockerComposeNetwork{
			Subnet: subnet,
		},
	}

	buf := new(bytes.Buffer)
	tpl := template.Must(template.New("docker-compose").Parse(string(dockerComposeTpl)))
	orPanic(tpl.Execute(buf, cfg))
	orPanic(os.WriteFile(outDir+"/docker-compose.yml", buf.Bytes(), 0o644))
}
