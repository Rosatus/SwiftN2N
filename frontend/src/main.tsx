import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

const container = document.getElementById('root')

class ErrorBoundary extends React.Component<{ children: React.ReactNode }, { error: string }> {
    constructor(props: { children: React.ReactNode }) {
        super(props)
        this.state = {error: ''}
    }

    static getDerivedStateFromError(error: unknown) {
        return {error: error instanceof Error ? error.message : String(error)}
    }

    componentDidCatch(error: unknown) {
        console.error(error)
    }

    render() {
        if (this.state.error) {
            return (
                <main className="shell">
                    <section className="fatal-panel">
                        <span className="section-kicker">Runtime error</span>
                        <h1>SwiftN2N could not render</h1>
                        <pre>{this.state.error}</pre>
                    </section>
                </main>
            )
        }
        return this.props.children
    }
}

window.addEventListener('error', (event) => {
    renderFatal('Runtime error', event.message)
})

window.addEventListener('unhandledrejection', (event) => {
    renderFatal('Unhandled rejection', String(event.reason))
})

if (!container) {
    throw new Error('Root element #root was not found')
}

const root = createRoot(container)

root.render(
    <React.StrictMode>
        <ErrorBoundary>
            <App/>
        </ErrorBoundary>
    </React.StrictMode>
)

window.setTimeout(() => {
    const rootElement = document.getElementById('root')
    const shell = rootElement?.querySelector('.shell') as HTMLElement | null
    const rect = shell?.getBoundingClientRect()
    if (!shell || !rect || rect.width < 20 || rect.height < 20) {
        renderFatal('Render watchdog', 'The React tree mounted without a visible shell. This usually means the WebView failed during startup.')
    }
}, 1600)

function renderFatal(label: string, message: string) {
    const rootElement = document.getElementById('root')
    if (!rootElement) {
        return
    }
    rootElement.innerHTML = `<main class="shell"><section class="fatal-panel"><span class="section-kicker">${escapeHtml(label)}</span><h1>SwiftN2N could not render</h1><pre>${escapeHtml(message)}</pre></section></main>`
}

function escapeHtml(value: string) {
    return value
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#039;')
}
