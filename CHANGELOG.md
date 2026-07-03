# Changelog

Historial de cambios de **Gentle-QE**, el fork de [gentle-ai](https://github.com/Gentleman-Programming/gentle-ai) orientado a QA/SDET.
Los cambios se listan en orden cronológico descendente (lo más reciente primero).

## [0.1.1] - 2026-07-03

- La persona ahora se comunica en español neutro (se eliminó el voseo rioplatense heredado del upstream).
- Homebrew instala con `brew install gentle-qe` (fórmula en lugar de cask; funciona también en Linux y sin pasos extra en macOS).
- Sincronización con la última versión del upstream gentle-ai: flujo SDD más liviano, mejoras del registro de skills y de la guía de codegraph.
- El historial del fork quedó realineado con el upstream: la comparación de GitHub ahora refleja la diferencia real y sirve como aviso de cuándo actualizar.
- Arreglado el CI: los tests E2E de respaldo buscaban el directorio de estado con el nombre viejo y hacían fallar los builds de `main`.
- Se agrega este changelog.

## [0.1.0] - 2026-07-03

Primer release público del fork.

- Rebrand completo desde gentle-ai: nuevo nombre en comandos, ayuda, TUI y directorio de estado (`~/.gentle-qe`).
- Persona SDET/QA Senior en español como identidad por defecto, con preset QE-SDET.
- Catálogo QA agregado sobre el upstream: 16 skills de testing (Playwright, accesibilidad, k6, contratos de API, etc.) y 8 agentes especializados.
- Mecanismo de overlay QE: los cambios del fork viven en una capa propia re-aplicable sobre el upstream, con verificación anti-drift local y en CI.
- Distribución multiplataforma: binarios para macOS/Linux/Windows, tap de Homebrew y bucket de Scoop.
- CI propio: tests unitarios, E2E en Ubuntu/Arch/Fedora y guard de integridad del overlay.

[0.1.1]: https://github.com/EduardoVeraE/gentle-qe/releases/tag/v0.1.1
[0.1.0]: https://github.com/EduardoVeraE/gentle-qe/releases/tag/v0.1.0
