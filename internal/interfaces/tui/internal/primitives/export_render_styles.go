package primitives

type RenderStyles struct {
	inner renderStyles
}

func NewRenderStyles(enabled bool) RenderStyles {
	return RenderStyles{inner: renderStyles{enabled: enabled}}
}

func ResolveRenderStylesFromEnv() RenderStyles {
	return RenderStyles{inner: resolveRenderStylesFromEnv()}
}

func (s RenderStyles) Enabled() bool {
	return s.inner.enabled
}

func (s RenderStyles) Title(text string) string {
	return s.inner.title(text)
}

func (s RenderStyles) Selected(text string) string {
	return s.inner.selected(text)
}

func (s RenderStyles) Muted(text string) string {
	return s.inner.muted(text)
}

func (s RenderStyles) Error(text string) string {
	return s.inner.error(text)
}

func (s RenderStyles) Dirty(text string) string {
	return s.inner.dirty(text)
}

func (s RenderStyles) Label(text string) string {
	return s.inner.label(text)
}

func (s RenderStyles) Summary(text string) string {
	return s.inner.summary(text)
}

func IsErrorLikeMessage(message string) bool {
	return isErrorLikeMessage(message)
}
