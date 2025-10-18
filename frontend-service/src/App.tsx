import React, { useEffect, useState } from 'react'
import './styles.css'
import axios from 'axios'

declare global {
  interface Window { Telegram: any }
}

type Expense = {
  id: number
  amount_cents: number
  category_id?: number | null
  timestamp: string
  is_shared: boolean
}

const API_BASE = '/api'

const App: React.FC = () => {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [amount, setAmount] = useState<string>('')
  const [profile, setProfile] = useState<{username?:string,id?:string,photo_url?:string}|null>(() => {
    try { return JSON.parse(localStorage.getItem('profile') || 'null') } catch { return null }
  })

  // We will dynamically inject the Telegram widget only when the login UI is shown
  const widgetRef = React.useRef<HTMLDivElement | null>(null)
  const [widgetLoading, setWidgetLoading] = useState<boolean>(false)
  const [widgetError, setWidgetError] = useState<string | null>(null)

  useEffect(() => {
    // If user is authenticated, ensure widget is not present
    if (token) {
      if (widgetRef.current) widgetRef.current.innerHTML = ''
      return
    }

    // create container for widget
    const container = widgetRef.current
    if (!container) return

    setWidgetLoading(true)
    setWidgetError(null)

    // prepare global callback that will forward the payload into React
    const win: any = window
    win.TelegramLoginWidget = win.TelegramLoginWidget || {}
    win.TelegramLoginWidget.onAuth = (user: Record<string, any>) => {
      // call React handler directly
      void onTelegramAuth(user)
    }

    // inject the widget script into the container
    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?15'
    script.async = true
    script.setAttribute('data-telegram-login', 'rd_expense_tracker_bot')
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-userpic', 'false')
    script.setAttribute('data-lang', 'en')
    script.setAttribute('data-onauth', 'window.TelegramLoginWidget.onAuth')

    // handle load/error events
    const onLoad = () => { setWidgetLoading(false); setWidgetError(null) }
    const onError = (err: any) => { setWidgetLoading(false); setWidgetError('Failed to load Telegram widget'); console.warn('widget load error', err) }
    script.addEventListener('load', onLoad)
    script.addEventListener('error', onError)

    container.appendChild(script)

    return () => {
      // cleanup
      try { container.innerHTML = '' } catch (e) {}
      try { delete (window as any).TelegramLoginWidget.onAuth } catch (e) {}
      script.removeEventListener('load', onLoad)
      script.removeEventListener('error', onError)
    }
  }, [token])

  useEffect(() => {
    // Listen for the custom event forwarded from the index.html widget
    const handler = (e: any) => {
      // event detail contains the Telegram payload
      const ev = e as CustomEvent<Record<string, any>>
      if (ev && ev.detail) onTelegramAuth(ev.detail)
    }
    window.addEventListener('telegramAuth', handler as EventListener)
    return () => window.removeEventListener('telegramAuth', handler as EventListener)
  }, [])

  useEffect(() => {
    if (token) fetchExpenses()
  }, [token])

  const onTelegramAuth = async (user: Record<string, any>) => {
    // user contains id, first_name, username, auth_date, hash, etc.
    try {
      const res = await axios.post(`${API_BASE}/login`, user)
  const t = res.data.token
  const p: any = { username: res.data.username, id: res.data.id }
  if (res.data.photo_url) p.photo_url = res.data.photo_url
  localStorage.setItem('token', t)
  localStorage.setItem('profile', JSON.stringify(p))
  setProfile(p)
  setToken(t)
  fetchExpenses(t)
    } catch (err) {
      alert('Login failed: ' + String(err))
    }
  }

  // Development fallback: simulate widget payload locally (client-side only)
  // This avoids calling the real /login endpoint (which requires a valid Telegram hash)
  const simulateLogin = async () => {
    const demoId = String(123456789)
    const demoProfile = { username: 'demo', id: demoId }
    const demoToken = 'dev-token-' + demoId
    // store locally (no server-side verification)
    localStorage.setItem('token', demoToken)
    localStorage.setItem('profile', JSON.stringify(demoProfile))
    setProfile(demoProfile)
    setToken(demoToken)
    // clear expenses cache for dev session
    setExpenses([])
  }

  const fetchExpenses = async (t?: string) => {
    try {
      const headers = t ? { Authorization: `Bearer ${t}` } : (token ? { Authorization: `Bearer ${token}` } : undefined)
      const res = await axios.get(`${API_BASE}/expenses`, { headers })
      setExpenses(res.data)
    } catch (err) {
      console.error('fetchExpenses', err)
    }
  }

  const addExpense = async () => {
    const cents = Math.round(parseFloat(amount || '0') * 100)
    if (!token) { alert('not authenticated'); return }
    try {
      await axios.post(`${API_BASE}/expenses`, { amount_cents: cents, timestamp: new Date().toISOString() }, { headers: { Authorization: `Bearer ${token}` } })
      setAmount('')
      fetchExpenses()
    } catch (err) {
      console.error('addExpense', err)
      alert('add failed')
    }
  }

  const logout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setExpenses([])
  }

  return (
    <div className="app-center">
      <div className="glass-card hero">
        <div>
          <h1 className="title">Expense Tracker</h1>
          <div className="subtitle">Minimal family expense tracker — login with Telegram to sync</div>
          {profile && (
            <div style={{ marginTop: 8, fontSize: 13, color: 'var(--muted)', display: 'flex', gap: 8, alignItems: 'center' }}>
              {profile.photo_url ? (
                <img src={profile.photo_url} alt="avatar" style={{ width:24,height:24,borderRadius:12,objectFit:'cover' }} />
              ) : null}
              <div>Signed in as <strong>{profile.username}</strong></div>
            </div>
          )}
        </div>
        <div className="controls">
          {!token ? (
            <div>
              <div>
                <div id="telegram-login-placeholder" ref={widgetRef} />
                {widgetLoading && <div className="spinner" aria-hidden />}
                {widgetError && (
                  <div className="widget-error">
                    <div>{widgetError}</div>
                    <div className="widget-retry">
                      <button onClick={() => {
                        // retry by clearing container and toggling token (force reload effect)
                        if (widgetRef.current) widgetRef.current.innerHTML = ''
                        setWidgetLoading(true)
                        setWidgetError(null)
                        // re-run effect: no-op, effect watches token only; we can trigger reload by toggling a ref key
                        setTimeout(() => {
                          if (widgetRef.current) {
                            const evt = new Event('widget-retry')
                            window.dispatchEvent(evt)
                          }
                        }, 50)
                      }}>Retry</button>
                      <button className="secondary" onClick={() => setWidgetError(null)}>Dismiss</button>
                    </div>
                  </div>
                )}
              </div>
              <button onClick={simulateLogin} className="secondary">Simulate Login (dev)</button>
            </div>
          ) : (
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <button onClick={logout} className="secondary">Logout</button>
            </div>
          )}
        </div>
      </div>

      <div className="glass-card">
        {!token ? (
          <div>
            <p className="subtitle">Sign in with Telegram to manage expenses. The widget below will allow signing in.</p>
          </div>
        ) : (
          <div>
            <div className="controls" style={{ marginBottom: 8 }}>
              <input type="text" placeholder="Amount (e.g. 12.34)" value={amount} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAmount(e.target.value)} />
              <button onClick={addExpense}>Add expense</button>
            </div>

            <h3 style={{ marginTop: 16 }}>Recent expenses</h3>
            <ul className="expenses">
              {expenses.map(e => (
                <li key={e.id}>{new Date(e.timestamp).toLocaleString()} — {(e.amount_cents/100).toFixed(2)} — shared: {String(e.is_shared)}</li>
              ))}
            </ul>
          </div>
        )}
      </div>

    </div>
  )
}

export default App
