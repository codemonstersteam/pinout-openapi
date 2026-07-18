# SpecLoader is built late via a Deps factory, not pre-built in Deps

The frozen design required the `SpecLoader` to be **both** bounded by `settings.timeout` **and**
encapsulated + constructed once as part of a `Deps` built entirely before `ProcessValidate` runs — but
`settings.timeout` is only known after `ConfigStore.Load`+`NewConfig` execute *inside* the pipe, so the
eager construction was forced to hardcode `defaultProviderTimeoutSeconds = 30`, leaving `settings.timeout`
inert and component scenario 6 unable ever to emit `TIMEOUT_ERROR`. We resolve it by making `Deps` carry a
**`BuildSpecLoader func(Settings) SpecLoader` factory** (the wired-once, encapsulated object) and
constructing the `SpecLoader` **once per invocation inside the head, after `NewConfig`**, bounded by the
real `cfg.Settings.Timeout`; the rejected alternative — making `main`/`register.go` read+validate the
config before building `Deps` — was refused because it duplicates config loading outside the pipe and
splits the composition root. `provider.NewSpecLoader(timeout)` keeps its signature; only its call site
moves. Reversing this is a `Deps`-shape + head-pipe change, hence recorded.
