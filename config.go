package mauerspecht

type Config struct {
	Hostname     string   `json:"hostname"`
	HTTPPorts    []int    `json:"http-ports"`
	MagicStrings []string `json:"magic-strings"`
}
