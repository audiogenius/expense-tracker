import React, { useEffect, useState } from 'react'
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

// Components
import { TelegramLogin } from './components/Auth/TelegramLogin'
import { Header } from './components/Header/Header'
import { BalanceCard } from './components/Balance/BalanceCard'
import { ExpenseLineChart } from './components/Charts/ExpenseLineChart'
import { CategoryPieChart } from './components/Charts/CategoryPieChart'
import { RecentTransactions } from './components/Transactions/RecentTransactions'
import { AddTransaction } from './components/Transactions/AddTransaction'
import { TransactionsPage } from './components/Transactions/TransactionsPage'
import { CategoriesPage } from './components/Categories/CategoriesPage'
import { SettingsPage } from './components/Settings/SettingsPage'

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
        <TelegramLogin onAuth={handleTelegramAuth} />
      </div>
    )
  }

  if (currentPage === 'transactions') {
    return (
      <div className="app-center">
        <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />
        <div className="page-header">
          <button onClick={handleBackToDashboard} className="back-btn">
            ← Назад к дашборду
          </button>
        </div>
        <TransactionsPage token={token!} />
      </div>
    )
  }

  if (currentPage === 'categories') {
    return (
      <div className="app-center">
        <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />
        <div className="page-header">
          <button onClick={handleBackToDashboard} className="back-btn">
            ← Назад к дашборду
          </button>
        </div>
        <CategoriesPage token={token!} />
      </div>
    )
  }

  if (currentPage === 'settings') {
    return (
      <div className="app-center">
        <Header profile={profile} onLogout={handleLogout} />
        <SettingsPage token={token!} onBack={handleBackToDashboard} />
      </div>
    )
  }

  return (
    <div className="app-center">
      <Header profile={profile} onLogout={handleLogout} onSettings={handleViewSettings} />

      <div className="main-grid">
        {/* LEFT COLUMN */}
        <div className="left-column">
          <BalanceCard 
            balance={balance} 
            filterPeriod={filterPeriod} 
            customPeriod={customPeriod}
            onPeriodChange={setFilterPeriod}
            onCustomPeriodChange={setCustomPeriod}
          />

          {/* Charts */}
          {(filteredExpenses.length > 0 || incomes.length > 0) && (
            <div className="charts-grid">
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
            </div>
          )}
        </div>

        {/* RIGHT COLUMN */}
        <div className="right-column">
          <RecentTransactions 
            token={token!} 
            onViewAll={handleViewAllTransactions}
          />
          <AddTransaction token={token!} onTransactionAdded={handleTransactionAdded} />
        </div>
      </div>
    </div>
  )
}

export default App
