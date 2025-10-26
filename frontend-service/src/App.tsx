import React, { useEffect, useState, Suspense, lazy } from 'react'
import './styles.css'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js'

// Types
import type { Expense, Income, Category, Balance, Profile, Period, CustomPeriod, ChartType, ChartPeriod } from './types'

// API
import * as api from './api'

// Lazy loaded components
const TelegramLogin = lazy(() => import('./components/Auth/TelegramLogin').then(m => ({ default: m.TelegramLogin })))
const Header = lazy(() => import('./components/Header/Header').then(m => ({ default: m.Header })))
const BalanceCard = lazy(() => import('./components/Balance/BalanceCard').then(m => ({ default: m.BalanceCard })))
const ExpenseLineChart = lazy(() => import('./components/Charts/ExpenseLineChart').then(m => ({ default: m.ExpenseLineChart })))
const CategoryPieChart = lazy(() => import('./components/Charts/CategoryPieChart').then(m => ({ default: m.CategoryPieChart })))
const RecentTransactions = lazy(() => import('./components/Transactions/RecentTransactions').then(m => ({ default: m.RecentTransactions })))
const AddTransaction = lazy(() => import('./components/Transactions/AddTransaction').then(m => ({ default: m.AddTransaction })))
const TransactionsPage = lazy(() => import('./components/Transactions/TransactionsPage').then(m => ({ default: m.TransactionsPage })))
const CategoriesPage = lazy(() => import('./components/Categories/CategoriesPage').then(m => ({ default: m.CategoriesPage })))
const SettingsPage = lazy(() => import('./components/Settings/SettingsPage').then(m => ({ default: m.SettingsPage })))

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, ArcElement)

const App: React.FC = () => {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [incomes, setIncomes] = useState<Income[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [balance, setBalance] = useState<Balance | null>(null)
  const [profile, setProfile] = useState<Profile | null>(() => {
    try {
      return JSON.parse(localStorage.getItem('profile') || 'null')
    } catch {
      return null
    }
  })

  // Filters
  const [filterCategory, setFilterCategory] = useState<number | null>(null)
  const [filterPeriod, setFilterPeriod] = useState<Period>('all')
  const [customPeriod, setCustomPeriod] = useState<CustomPeriod | undefined>(undefined)
  const [currentPage, setCurrentPage] = useState<'dashboard' | 'transactions' | 'categories' | 'settings'>('dashboard')
  
  // Chart settings
  const [chartType, setChartType] = useState<ChartType>('expenses')
  const [chartPeriod, setChartPeriod] = useState<ChartPeriod>('days')
  
  // Refresh trigger for RecentTransactions
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  // Fetch data on mount and period change
  useEffect(() => {
    if (token) {
      fetchData()
    }
  }, [token, filterPeriod, customPeriod])

  const fetchData = async () => {
    if (!token) return
    try {
      const [expensesData, incomesData, categoriesData, balanceData] = await Promise.all([
        api.fetchExpenses(token),
        api.fetchIncomes(token),
        api.fetchCategories(),
        api.fetchBalance(token, filterPeriod, customPeriod)
      ])
      setExpenses(expensesData)
      setIncomes(incomesData)
      setCategories(categoriesData)
      setBalance(balanceData)
    } catch (err) {
      console.error('fetchData error', err)
    }
  }

  const handleTelegramAuth = async (authData: Record<string, any>) => {
    try {
      const { token: newToken, profile: newProfile } = await api.loginWithTelegram(authData)
      localStorage.setItem('token', newToken)
      localStorage.setItem('profile', JSON.stringify(newProfile))
      setProfile(newProfile)
      setToken(newToken)
    } catch (err) {
      alert('Login failed: ' + String(err))
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('profile')
    setToken(null)
    setExpenses([])
    setIncomes([])
    setProfile(null)
    setBalance(null)
  }


  const handleTransactionAdded = async () => {
    await fetchData()
    setRefreshTrigger(prev => prev + 1)
  }

  const handleViewAllTransactions = () => {
    setCurrentPage('transactions')
  }

  const handleViewCategories = () => {
    setCurrentPage('categories')
  }

  const handleViewSettings = () => {
    setCurrentPage('settings')
  }

  const handleBackToDashboard = () => {
    setCurrentPage('dashboard')
  }

  // Filter expenses
  const filteredExpenses = expenses.filter((e) => {
    if (filterCategory !== null && e.category_id !== filterCategory) return false

    if (filterPeriod !== 'all') {
      const expenseDate = new Date(e.timestamp)
      const now = new Date()

      if (filterPeriod === 'day') {
        const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
        if (expenseDate < today) return false
      } else if (filterPeriod === 'week') {
        const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        if (expenseDate < weekAgo) return false
      } else if (filterPeriod === 'month') {
        const monthAgo = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate())
        if (expenseDate < monthAgo) return false
      } else if (filterPeriod === 'custom' && customPeriod) {
        const startDate = new Date(customPeriod.start_date)
        const endDate = new Date(customPeriod.end_date)
        endDate.setHours(23, 59, 59, 999)
        if (expenseDate < startDate || expenseDate > endDate) return false
      }
    }

    return true
  })

  if (!token) {
    return (
      <div className="app-center">
        <Suspense fallback={<div className="loading-spinner">Загрузка входа...</div>}>
          <TelegramLogin onAuth={handleTelegramAuth} />
        </Suspense>
      </div>
    )
  }

  if (currentPage === 'transactions') {
    return (
      <div className="app-center">
        <Suspense fallback={<div className="loading-spinner">Загрузка...</div>}>
          <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />
        </Suspense>
        <div className="page-header">
          <button onClick={handleBackToDashboard} className="back-btn">
            ← Назад к дашборду
          </button>
        </div>
        <Suspense fallback={<div className="loading-spinner">Загрузка транзакций...</div>}>
          <TransactionsPage token={token!} />
        </Suspense>
      </div>
    )
  }

  if (currentPage === 'categories') {
    return (
      <div className="app-center">
        <Suspense fallback={<div className="loading-spinner">Загрузка...</div>}>
          <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />
        </Suspense>
        <div className="page-header">
          <button onClick={handleBackToDashboard} className="back-btn">
            ← Назад к дашборду
          </button>
        </div>
        <Suspense fallback={<div className="loading-spinner">Загрузка категорий...</div>}>
          <CategoriesPage token={token!} />
        </Suspense>
      </div>
    )
  }

  if (currentPage === 'settings') {
    return (
      <div className="app-center">
        <Suspense fallback={<div className="loading-spinner">Загрузка...</div>}>
          <Header profile={profile} onLogout={handleLogout} />
        </Suspense>
        <Suspense fallback={<div className="loading-spinner">Загрузка настроек...</div>}>
          <SettingsPage token={token!} onBack={handleBackToDashboard} />
        </Suspense>
      </div>
    )
  }

  return (
    <div className="app-center">
      <Suspense fallback={<div className="loading-spinner">Загрузка...</div>}>
        <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />
      </Suspense>

      <div className="main-grid">
        {/* LEFT COLUMN */}
        <div className="left-column">
          <Suspense fallback={<div className="loading-spinner">Загрузка баланса...</div>}>
            <BalanceCard 
              balance={balance} 
              filterPeriod={filterPeriod} 
              customPeriod={customPeriod}
              onPeriodChange={setFilterPeriod}
              onCustomPeriodChange={setCustomPeriod}
            />
          </Suspense>

          {/* Charts */}
          {(filteredExpenses.length > 0 || incomes.length > 0) && (
            <div className="charts-grid">
              <Suspense fallback={<div className="loading-spinner">Загрузка графиков...</div>}>
                <ExpenseLineChart 
                  expenses={filteredExpenses} 
                  incomes={incomes}
                  chartType={chartType}
                  chartPeriod={chartPeriod}
                  onChartTypeChange={setChartType}
                  onChartPeriodChange={setChartPeriod}
                />
                {filteredExpenses.length > 0 && (
                  <CategoryPieChart expenses={filteredExpenses} categories={categories} />
                )}
              </Suspense>
            </div>
          )}
        </div>

        {/* RIGHT COLUMN */}
        <div className="right-column">
          <Suspense fallback={<div className="loading-spinner">Загрузка транзакций...</div>}>
            <RecentTransactions 
              token={token!} 
              onViewAll={handleViewAllTransactions}
              refreshTrigger={refreshTrigger}
            />
          </Suspense>
          <Suspense fallback={<div className="loading-spinner">Загрузка формы...</div>}>
            <AddTransaction token={token!} onTransactionAdded={handleTransactionAdded} />
          </Suspense>
        </div>
      </div>
    </div>
  )
}

export default App
