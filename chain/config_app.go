package chain

import (
	"bytes"
	"os"
	"text/template"

	_ "embed"
)

type AppConfig struct {
	MinimumGasPrices string
	EVMEnabled       bool
	ProdLike         bool
}

func (appConfig *AppConfig) Save(homeDir string) {
	orPanic(os.MkdirAll(homeDir+"/config", 0o700))

	if len(appConfig.MinimumGasPrices) == 0 {
		appConfig.MinimumGasPrices = "0" + DefaultBondDenom
	}

	buf := new(bytes.Buffer)

	if appConfig.EVMEnabled {
		tpl := template.Must(template.New("app_evm").Parse(string(appEvmTplTOML)))
		orPanic(tpl.Execute(buf, appConfig))
	} else if appConfig.ProdLike {
		tpl := template.Must(template.New("app_prod").Parse(string(appProdTplTOML)))
		orPanic(tpl.Execute(buf, appConfig))
	} else {
		tpl := template.Must(template.New("app").Parse(string(appTplTOML)))
		orPanic(tpl.Execute(buf, appConfig))
	}

	orPanic(os.WriteFile(homeDir+"/config/app.toml", buf.Bytes(), 0o600))
}

var (
	//go:embed templates/app.toml.tpl
	appTplTOML []byte

	//go:embed templates/app.evm.toml.tpl
	appEvmTplTOML []byte

	//go:embed templates/app.prod.toml.tpl
	appProdTplTOML []byte
)
