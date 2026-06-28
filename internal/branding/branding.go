// Package branding centraliza la identidad de producto del fork Gentle-QE.
//
// El fork mantiene un rebranding MÍNIMO sobre el upstream gentleman-programming/gentle-ai:
// solo se rebrandean los puntos funcionalmente necesarios (self-update, directorio de
// estado, version string, instalador). El module path Go NO se rebrandea.
//
// Centralizar estos valores aquí reduce la fricción de sync: si un merge del upstream
// re-introduce un literal "gentle-ai" en un sitio de anclaje, la herramienta
// tools/qe-overlay lo detecta y la corrección es trivial (apuntar a branding.*).
package branding

const (
	// Product es el nombre del binario/producto del fork.
	Product = "gentle-qe"
	// Display es el nombre visible de la marca.
	Display = "Gentle-QE"
	// Owner es el owner del repositorio del fork (releases, self-update).
	Owner = "EduardoVeraE"
	// Repo es el nombre del repositorio del fork.
	Repo = "gentle-qe"
	// StateDir es el directorio de estado/configuración bajo el home del usuario.
	StateDir = ".gentle-qe"
	// UserAgent es el User-Agent usado en las llamadas de chequeo de actualización.
	UserAgent = "gentle-qe-update-check"
)
