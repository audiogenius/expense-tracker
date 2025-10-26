import React, { useEffect, useState, useRef } from 'react'

type TelegramLoginProps = {
  onAuth: (authData: Record<string, any>) => void
}

export const TelegramLogin = ({ onAuth }: TelegramLoginProps) => {
  const widgetRef = useRef<HTMLDivElement | null>(null)
  const [widgetLoading, setWidgetLoading] = useState<boolean>(false)
  const [widgetError, setWidgetError] = useState<string | null>(null)
  const [domainInvalid, setDomainInvalid] = useState<boolean>(false)

  // Check for Telegram auth params in URL on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    if (params.has('id') && params.has('hash')) {
      const authData: Record<string, string> = {}
      params.forEach((value, key) => {
        authData[key] = value
      })

      // Clean URL
      window.history.replaceState({}, document.title, window.location.pathname)

      // Process auth
      onAuth(authData)
    }
  }, [onAuth])

  // Listen for "Bot domain invalid" error from Telegram iframe
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data && typeof event.data === 'string') {
        if (event.data.includes('Bot domain invalid') || event.data.includes('domain')) {
          setDomainInvalid(true)
        }
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  // Inject Telegram widget
  useEffect(() => {
    const container = widgetRef.current
    if (!container) return

    setWidgetLoading(true)
    setWidgetError(null)

    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.setAttribute('data-telegram-login', 'rd_expense_tracker_bot')
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-userpic', 'false')
    script.setAttribute('data-request-access', 'write')
    script.setAttribute('data-lang', 'ru')
    script.setAttribute('data-auth-url', window.location.origin + window.location.pathname)

    const onLoad = () => {
      setWidgetLoading(false)
      setWidgetError(null)
    }
    const onError = (err: any) => {
      setWidgetLoading(false)
      setWidgetError('Failed to load Telegram widget')
      console.warn('widget load error', err)
    }
    script.addEventListener('load', onLoad)
    script.addEventListener('error', onError)

    container.appendChild(script)

    return () => {
      try {
        container.innerHTML = ''
      } catch (e) {}
      script.removeEventListener('load', onLoad)
      script.removeEventListener('error', onError)
    }
  }, [])

  // Show "НЕТ" fullscreen if domain is invalid
  if (domainInvalid) {
    return (
      <div
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--bg)',
          zIndex: 9999
        }}
      >
        <div
          style={{
            fontSize: 'clamp(120px, 30vw, 300px)',
            fontWeight: 900,
            color: '#ffffff',
            textAlign: 'center',
            lineHeight: 1,
            letterSpacing: '-0.05em',
            animation: 'pulse 2s ease-in-out infinite',
            textShadow: '0 0 40px rgba(255, 255, 255, 0.3)'
          }}
        >
          НЕТ
        </div>
      </div>
    )
  }

  return (
    <div className="glass-card hero">
      <div className="hero-content">
        <h1 className="title">Expense Tracker</h1>
        <div className="subtitle">Минимальный семейный трекер расходов — вход через Telegram</div>
        <div id="telegram-login-placeholder" ref={widgetRef} />
        {widgetLoading && <div className="spinner" aria-hidden />}
        {widgetError && (
          <div className="widget-error">
            <div>{widgetError}</div>
            <div className="widget-retry">
              <button
                onClick={() => {
                  if (widgetRef.current) widgetRef.current.innerHTML = ''
                  setWidgetLoading(true)
                  setWidgetError(null)
                  setTimeout(() => {
                    if (widgetRef.current) {
                      const evt = new Event('widget-retry')
                      window.dispatchEvent(evt)
                    }
                  }, 50)
                }}
              >
                Retry
              </button>
              <button className="secondary" onClick={() => setWidgetError(null)}>
                Dismiss
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

