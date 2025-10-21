import React, { useEffect, useState } from 'react'
import './styles.css'
import axios from 'axios'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
} from 'chart.js'
import { Line, Pie } from 'react-chartjs-2'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
)

declare global {
  interface Window { Telegram: any }
}

type Expense = {
  id: number
  user_id: number
  amount_cents: number
  category_id?: number | null
  timestamp: string
  is_shared: boolean
  username?: string
}

type Income = {
  id: number
  user_id: number
  amount_cents: number
  income_type: string
  description?: string
  related_debt_id?: number | null
  timestamp: string
  username?: string
}

type Category = {
  id: number
  name: string
  aliases: string[]
}

type Balance = {
  balance_cents: number
  balance_rubles: number
  total_incomes_cents: number
  total_incomes_rubles: number
  total_expenses_cents: number
  total_expenses_rubles: number
  period: string
}

const API_BASE = '/api'

const App: React.FC = () => {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [incomes, setIncomes] = useState<Income[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [amount, setAmount] = useState<string>('')
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)
  const [profile, setProfile] = useState<{username?:string,id?:string,photo_url?:string}|null>(() => {
    try { return JSON.parse(localStorage.getItem('profile') || 'null') } catch { return null }
  })

  // Incomes state
  const [incomeAmount, setIncomeAmount] = useState<string>('')
  const [incomeType, setIncomeType] = useState<string>('salary')
  const [incomeDescription, setIncomeDescription] = useState<string>('')

  // Filters
  const [filterCategory, setFilterCategory] = useState<number | null>(null)
  const [filterPeriod, setFilterPeriod] = useState<'all' | 'week' | 'month'>('all')
  const [totalExpenses, setTotalExpenses] = useState<{total_cents: number, total_rubles: number}>({total_cents: 0, total_rubles: 0})
  const [balance, setBalance] = useState<Balance | null>(null)

  // We will dynamically inject the Telegram widget only when the login UI is shown
  const widgetRef = React.useRef<HTMLDivElement | null>(null)
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
      void onTelegramAuth(authData)
    }
  }, [])

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

    // inject the widget script into the container
    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.setAttribute('data-telegram-login', 'rd_expense_tracker_bot')
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-userpic', 'false')
    script.setAttribute('data-request-access', 'write')
    script.setAttribute('data-lang', 'ru')
    script.setAttribute('data-auth-url', window.location.origin + window.location.pathname)

    // handle load/error events
    const onLoad = () => { setWidgetLoading(false); setWidgetError(null) }
    const onError = (err: any) => { setWidgetLoading(false); setWidgetError('Failed to load Telegram widget'); console.warn('widget load error', err) }
    script.addEventListener('load', onLoad)
    script.addEventListener('error', onError)

    container.appendChild(script)

    return () => {
      // cleanup
      try { container.innerHTML = '' } catch (e) {}
      script.removeEventListener('load', onLoad)
      script.removeEventListener('error', onError)
    }
  }, [token])

  useEffect(() => {
    if (token) {
      fetchExpenses()
      fetchIncomes()
      fetchCategories()
      fetchTotalExpenses(filterPeriod)
      fetchBalance(filterPeriod)
    }
  }, [token, filterPeriod])

  const onTelegramAuth = async (user: Record<string, any>) => {
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
      fetchCategories()
    } catch (err) {
      alert('Login failed: ' + String(err))
    }
  }


  const fetchCategories = async () => {
    try {
      const res = await axios.get(`${API_BASE}/categories`)
      setCategories(res.data || [])
    } catch (err) {
      console.error('fetchCategories', err)
    }
  }

  const fetchExpenses = async (t?: string) => {
    try {
      const headers = t ? { Authorization: `Bearer ${t}` } : (token ? { Authorization: `Bearer ${token}` } : undefined)
      const res = await axios.get(`${API_BASE}/expenses`, { headers })
      setExpenses(res.data || [])
    } catch (err) {
      console.error('fetchExpenses', err)
    }
  }

  const fetchTotalExpenses = async (period: string) => {
    try {
      const headers = token ? { Authorization: `Bearer ${token}` } : undefined
      const res = await axios.get(`${API_BASE}/expenses/total?period=${period}`, { headers })
      setTotalExpenses(res.data)
    } catch (err) {
      console.error('fetchTotalExpenses', err)
    }
  }

  const fetchIncomes = async (t?: string) => {
    try {
      const headers = t ? { Authorization: `Bearer ${t}` } : (token ? { Authorization: `Bearer ${token}` } : undefined)
      const res = await axios.get(`${API_BASE}/incomes`, { headers })
      setIncomes(res.data || [])
    } catch (err) {
      console.error('fetchIncomes', err)
    }
  }

  const fetchBalance = async (period: string) => {
    try {
      const headers = token ? { Authorization: `Bearer ${token}` } : undefined
      const res = await axios.get(`${API_BASE}/balance?period=${period}`, { headers })
      setBalance(res.data)
    } catch (err) {
      console.error('fetchBalance', err)
    }
  }

  const addExpense = async () => {
    const cents = Math.round(parseFloat(amount || '0') * 100)
    if (!token) { alert('not authenticated'); return }
    if (cents <= 0) { alert('Amount must be positive'); return }
    try {
      await axios.post(`${API_BASE}/expenses`, { 
        amount_cents: cents, 
        category_id: selectedCategory,
        timestamp: new Date().toISOString() 
      }, { headers: { Authorization: `Bearer ${token}` } })
      setAmount('')
      setSelectedCategory(null)
      fetchExpenses()
      fetchTotalExpenses(filterPeriod)
      fetchBalance(filterPeriod)
    } catch (err) {
      console.error('addExpense', err)
      alert('Add failed')
    }
  }

  const addIncome = async () => {
    const cents = Math.round(parseFloat(incomeAmount || '0') * 100)
    if (!token) { alert('not authenticated'); return }
    if (cents <= 0) { alert('Сумма должна быть положительной'); return }
    try {
      await axios.post(`${API_BASE}/incomes`, { 
        amount_cents: cents, 
        income_type: incomeType,
        description: incomeDescription || '',
        timestamp: new Date().toISOString() 
      }, { headers: { Authorization: `Bearer ${token}` } })
      setIncomeAmount('')
      setIncomeType('salary')
      setIncomeDescription('')
      fetchIncomes()
      fetchBalance(filterPeriod)
    } catch (err) {
      console.error('addIncome', err)
      alert('Ошибка добавления прихода')
    }
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('profile')
    setToken(null)
    setExpenses([])
    setIncomes([])
    setProfile(null)
    setBalance(null)
  }

  const getIncomeTypeName = (type: string) => {
    const types: Record<string, string> = {
      'salary': 'Зарплата',
      'debt_return': 'Возврат долга',
      'prize': 'Выигрыш',
      'gift': 'Подарок',
      'refund': 'Возврат средств',
      'other': 'Прочее'
    }
    return types[type] || 'Неизвестно'
  }

  // Filter expenses
  const filteredExpenses = expenses.filter(e => {
    if (filterCategory !== null && e.category_id !== filterCategory) return false
    
    if (filterPeriod !== 'all') {
      const expenseDate = new Date(e.timestamp)
      const now = new Date()
      const daysDiff = (now.getTime() - expenseDate.getTime()) / (1000 * 60 * 60 * 24)
      
      if (filterPeriod === 'week' && daysDiff > 7) return false
      if (filterPeriod === 'month' && daysDiff > 30) return false
    }
    
    return true
  })

  // Prepare chart data
  const prepareLineChartData = () => {
    const last7Days = [...Array(7)].map((_, i) => {
      const d = new Date()
      d.setDate(d.getDate() - (6 - i))
      return d.toISOString().split('T')[0]
    })

    const expensesByDay = last7Days.map(day => {
      return filteredExpenses
        .filter(e => e.timestamp.split('T')[0] === day)
        .reduce((sum, e) => sum + e.amount_cents, 0) / 100
    })

    return {
      labels: last7Days.map(d => new Date(d).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })),
      datasets: [{
        label: 'Расходы (₽)',
        data: expensesByDay,
        borderColor: 'rgb(124, 58, 237)',
        backgroundColor: 'rgba(124, 58, 237, 0.1)',
        tension: 0.3,
      }]
    }
  }

  const preparePieChartData = () => {
    const expensesByCategory: Record<string, number> = {}
    
    filteredExpenses.forEach(e => {
      const cat = categories.find(c => c.id === e.category_id)
      const catName = cat ? cat.name : 'Без категории'
      expensesByCategory[catName] = (expensesByCategory[catName] || 0) + e.amount_cents / 100
    })

    const colors = [
      'rgba(124, 58, 237, 0.8)',
      'rgba(59, 130, 246, 0.8)',
      'rgba(16, 185, 129, 0.8)',
      'rgba(245, 158, 11, 0.8)',
      'rgba(239, 68, 68, 0.8)',
      'rgba(168, 85, 247, 0.8)',
      'rgba(6, 182, 212, 0.8)',
      'rgba(251, 146, 60, 0.8)',
    ]

    return {
      labels: Object.keys(expensesByCategory),
      datasets: [{
        data: Object.values(expensesByCategory),
        backgroundColor: colors,
      }]
    }
  }

  const getCategoryName = (categoryId: number | null | undefined) => {
    if (!categoryId) return 'Без категории'
    const cat = categories.find(c => c.id === categoryId)
    return cat ? cat.name : 'Неизвестно'
  }

  // Show "НЕТ" fullscreen if domain is invalid
  if (domainInvalid && !token) {
    return (
      <div style={{
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
      }}>
        <div style={{
          fontSize: 'clamp(120px, 30vw, 300px)',
          fontWeight: 900,
          color: '#ffffff',
          textAlign: 'center',
          lineHeight: 1,
          letterSpacing: '-0.05em',
          animation: 'pulse 2s ease-in-out infinite',
          textShadow: '0 0 40px rgba(255, 255, 255, 0.3)'
        }}>
          НЕТ
        </div>
      </div>
    )
  }

  return (
    <div className="app-center">
      {token ? (
        <>
          {/* Header with Profile and Logout */}
          <div className="header-bar">
            <div className="header-left">
              <h1 className="title" style={{ fontSize: '24px', margin: 0 }}>Expense Tracker</h1>
              {profile && (
                <div className="profile-avatar" style={{ padding: '8px 12px' }}>
                  {profile.photo_url && <img src={profile.photo_url} alt="avatar" style={{ width: '32px', height: '32px' }} />}
                  <div className="profile-info">
                    <div className="profile-name">{profile.username}</div>
                  </div>
                </div>
              )}
            </div>
            <button onClick={logout} className="logout-btn" title="Выход">✕</button>
          </div>
        </>
      ) : (
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
                  <button onClick={() => {
                    if (widgetRef.current) widgetRef.current.innerHTML = ''
                    setWidgetLoading(true)
                    setWidgetError(null)
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
        </div>
      )}

      {token && (
        <div className="main-grid">
          {/* LEFT COLUMN */}
          <div className="left-column">
            {/* Balance Card */}
            <div className="glass-card">
              <h3>Семейный бюджет</h3>
              <div className="period-buttons" style={{ marginBottom: 20 }}>
                <button 
                  className={filterPeriod === 'all' ? 'active' : 'secondary'}
                  onClick={() => setFilterPeriod('all')}
                >Все</button>
                <button 
                  className={filterPeriod === 'week' ? 'active' : 'secondary'}
                  onClick={() => setFilterPeriod('week')}
                >Неделя</button>
                <button 
                  className={filterPeriod === 'month' ? 'active' : 'secondary'}
                  onClick={() => setFilterPeriod('month')}
                >Месяц</button>
              </div>
              {balance && (
                <div className="balance-grid">
                  <div className="balance-card">
                    <div className="balance-label">Баланс</div>
                    <div className="balance-value" style={{ color: balance.balance_rubles >= 0 ? '#10b981' : '#ef4444' }}>
                      {balance.balance_rubles.toFixed(2)} ₽
                    </div>
                  </div>
                  <div className="balance-stats">
                    <div className="balance-stat">
                      <span className="balance-stat-label">Доходы:</span>
                      <span className="balance-stat-value" style={{ color: '#10b981' }}>
                        +{balance.total_incomes_rubles.toFixed(2)} ₽
                      </span>
                    </div>
                    <div className="balance-stat">
                      <span className="balance-stat-label">Расходы:</span>
                      <span className="balance-stat-value" style={{ color: '#ef4444' }}>
                        -{balance.total_expenses_rubles.toFixed(2)} ₽
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Charts */}
            {filteredExpenses.length > 0 && (
              <div className="charts-grid">
                <div className="glass-card chart-container">
                  <h3>Расходы за последние 7 дней</h3>
                  <Line data={prepareLineChartData()} options={{ 
                    responsive: true, 
                    maintainAspectRatio: true,
                    plugins: {
                      legend: {
                        labels: {
                          color: '#ffffff',
                          font: { size: 14, weight: 'bold' }
                        }
                      }
                    },
                    scales: {
                      x: {
                        ticks: { color: '#ffffff' },
                        grid: { color: 'rgba(255, 255, 255, 0.1)' }
                      },
                      y: {
                        ticks: { color: '#ffffff' },
                        grid: { color: 'rgba(255, 255, 255, 0.1)' }
                      }
                    }
                  }} />
                </div>
                <div className="glass-card chart-container">
                  <h3>Категории расходов</h3>
                  <Pie data={preparePieChartData()} options={{ 
                    responsive: true, 
                    maintainAspectRatio: true,
                    plugins: {
                      legend: {
                        labels: {
                          color: '#ffffff',
                          font: { size: 14, weight: 'bold' }
                        }
                      }
                    }
                  }} />
                </div>
              </div>
            )}
          </div>

          {/* RIGHT COLUMN */}
          <div className="right-column">
            {/* Expenses List */}
            <div className="glass-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20, flexWrap: 'wrap', gap: '12px' }}>
                <h3 style={{ margin: 0 }}>Последние расходы</h3>
                <select 
                  value={filterCategory || ''} 
                  onChange={(e) => setFilterCategory(e.target.value ? parseInt(e.target.value) : null)}
                >
                  <option value="">Все категории</option>
                  {categories.map(cat => (
                    <option key={cat.id} value={cat.id}>{cat.name}</option>
                  ))}
                </select>
              </div>
              <ul className="expenses">
                {filteredExpenses.length === 0 ? (
                  <li style={{ textAlign: 'center', color: 'var(--text-muted)', padding: '40px 20px' }}>Нет расходов</li>
                ) : (
                  filteredExpenses.map(e => (
                    <li key={e.id} className="expense-item">
                      <div className="expense-info">
                        <div className="expense-date">{new Date(e.timestamp).toLocaleString('ru-RU')}</div>
                        <div className="expense-category">{getCategoryName(e.category_id)}</div>
                      </div>
                      <div className="expense-amount">{(e.amount_cents/100).toFixed(2)} ₽</div>
                    </li>
                  ))
                )}
              </ul>
            </div>

            {/* Add Expense */}
            <div className="glass-card">
              <h3>Добавить расход</h3>
              <div className="controls">
                <input 
                  type="number" 
                  placeholder="Сумма (руб.)" 
                  value={amount} 
                  onChange={(e) => setAmount(e.target.value)} 
                />
                <select 
                  value={selectedCategory || ''} 
                  onChange={(e) => setSelectedCategory(e.target.value ? parseInt(e.target.value) : null)}
                >
                  <option value="">Без категории</option>
                  {categories.map(cat => (
                    <option key={cat.id} value={cat.id}>{cat.name}</option>
                  ))}
                </select>
                <button onClick={addExpense}>Добавить расход</button>
              </div>
            </div>

            {/* Add Income */}
            <div className="glass-card">
              <h3>Добавить приход</h3>
              <div className="controls">
                <input 
                  type="number" 
                  placeholder="Сумма (руб.)" 
                  value={incomeAmount} 
                  onChange={(e) => setIncomeAmount(e.target.value)} 
                />
                <select 
                  value={incomeType} 
                  onChange={(e) => setIncomeType(e.target.value)}
                >
                  <option value="salary">Зарплата</option>
                  <option value="debt_return">Возврат долга</option>
                  <option value="prize">Выигрыш</option>
                  <option value="gift">Подарок</option>
                  <option value="refund">Возврат средств</option>
                  <option value="other">Прочее</option>
                </select>
                <input 
                  type="text" 
                  placeholder="Описание (необязательно)" 
                  value={incomeDescription} 
                  onChange={(e) => setIncomeDescription(e.target.value)} 
                />
                <button onClick={addIncome} style={{ background: '#10b981' }}>Добавить приход</button>
              </div>
            </div>

            {/* Incomes List */}
            <div className="glass-card">
              <h3>Последние приходы</h3>
              <ul className="expenses">
                {incomes.length === 0 ? (
                  <li style={{ textAlign: 'center', color: 'var(--text-muted)', padding: '40px 20px' }}>Нет приходов</li>
                ) : (
                  incomes.slice(0, 20).map(inc => (
                    <li key={inc.id} className="expense-item">
                      <div className="expense-info">
                        <div className="expense-date">{new Date(inc.timestamp).toLocaleString('ru-RU')}</div>
                        <div className="expense-category">{getIncomeTypeName(inc.income_type)}</div>
                        {inc.description && <div className="income-description" style={{ fontSize: '13px', color: 'var(--text-muted)', marginTop: 4 }}>{inc.description}</div>}
                        {inc.username && <div className="income-user" style={{ fontSize: '12px', color: 'var(--accent)', marginTop: 2 }}>@{inc.username}</div>}
                      </div>
                      <div className="expense-amount" style={{ color: '#10b981' }}>+{(inc.amount_cents/100).toFixed(2)} ₽</div>
                    </li>
                  ))
                )}
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default App
