package meta

import (
	"net/url"
	"path"
	"runtime/debug"
	"strings"
)

var (
	NovuGoProviderVersion = "0.1.0"
	NovuGoRepoBaseUrl     = "https://github.com/novuhq/novu-go"
)

func init() {
	NovuGoProviderVersion = ActualModuleVersion("github.com/novuhq/novu-go")
}

func NovuGoDocsUrl(pth, ref string) string {
	if ref == "" {
		ref = NovuGoProviderVersion
	} else {
		ref = "main"
	}
	pth = strings.TrimPrefix(pth, "/")
	base := strings.TrimSuffix(NovuGoRepoBaseUrl, "/")

	u, err := url.Parse(base)
	if err != nil {
		//try to return something
		return base + "blob/" + ref + "/" + pth
	}
	u.Path = path.Join(u.Path, "blob", ref, pth)
	return u.String()
}

// Use the debug package to get the actualversion of the module
// warning, it could retourn a pseudo version if the module is not a release
func ActualModuleVersion(modulePath string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			return dep.Version
		}
	}
	return ""
}
