() => fetch('http://127.0.0.1:__PORT__/health', {signal: AbortSignal.timeout(2000)}).then((response) => response.ok, () => false)
